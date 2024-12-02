"""Square service."""

import asyncio
import itertools
from collections.abc import Callable, Iterator, Mapping
from typing import Any, Optional, ParamSpec, Type, TypeVar, cast

import nanoid
from apimatic_core.http.response.api_response import ApiResponse
from cattrs import Converter, override
from cattrs.gen import make_dict_unstructure_fn
from cattrs.preconf.orjson import make_converter
from loguru import logger
from oes.payment.currency import format_currency
from oes.payment.payment import (
    CancelPaymentRequest,
    CreatePaymentRequest,
    Payment,
    PaymentError,
    PaymentMethodConfig,
    PaymentMethodError,
    PaymentNotFoundError,
    PaymentResult,
    PaymentStateError,
    PaymentStatus,
    UpdatePaymentRequest,
)
from oes.payment.pricing import LineItem, Modifier, PricingResult
from oes.payment.services.square.models import (
    CreateCustomerBody,
    CreateOrder,
    CreateOrderLineItem,
    CreateOrderLineItemAppliedDiscount,
    CreateOrderLineItemDiscount,
    CreateOrderLineItemModifier,
    CreatePayment,
    CustomerResponse,
    CustomersResponse,
    Error,
    ErrorResponse,
    Money,
    Order,
    OrderLineItemDiscountScope,
    OrderLineItemDiscountType,
    OrderLineItemItemType,
    OrderResponseBody,
    OrderState,
    PaymentResponse,
    PricingOptions,
    SquareConfig,
)
from oes.payment.types import CartData, PaymentMethod
from square.client import Client, CustomersApi, OrdersApi, PaymentsApi

_P = ParamSpec("_P")
_T = TypeVar("_T")

_converter = make_converter()

# square client doesn't handle enums well...
_converter.register_unstructure_hook(
    CreateOrderLineItem,
    make_dict_unstructure_fn(
        CreateOrderLineItem,
        _converter,
        item_type=override(unstruct_hook=lambda v: v.value),
    ),
)
_converter.register_unstructure_hook(
    CreateOrderLineItemDiscount,
    make_dict_unstructure_fn(
        CreateOrderLineItemDiscount,
        _converter,
        type=override(unstruct_hook=lambda v: v.value),
        scope=override(unstruct_hook=lambda v: v.value),
    ),
)


class SquareService:
    """Square service."""

    id = "square"
    name = "Square"

    def __init__(
        self,
        application_id: str,
        access_token: str,
        location_id: str,
        sandbox: bool = False,
        item_map: Mapping[str, str] | None = None,
        modifier_map: Mapping[str, str] | None = None,
    ):
        self.application_id = application_id
        self.location_id = location_id
        self.environment = "production" if not sandbox else "sandbox"
        self._client = Client(
            access_token=access_token,
            environment=self.environment,
        )
        self._item_map = item_map or {}
        self._modifier_map = modifier_map or {}

    def get_payment_method(self, config: PaymentMethodConfig, /) -> PaymentMethod:
        """Get a :class:`PaymentMethod`."""
        return SquareWebPaymentMethod(self)

    async def cancel_payment(self, request: CancelPaymentRequest, /) -> PaymentResult:
        """Cancel a payment."""
        order = await self._get_order(request.external_id)

        if order.state != OrderState.open:
            raise PaymentStateError("Payment is already closed")

        res = await self._req(
            self._orders.update_order,
            OrderResponseBody,
            request.external_id,
            {"order": {"version": order.version, "state": OrderState.canceled.value}},
        )
        if isinstance(res, ErrorResponse):
            raise PaymentMethodError(_format_error(res))

        final_order = res.order

        return PaymentResult(
            id=request.id,
            service=self.id,
            external_id=final_order.id,
            status=_order_status_to_payment_status(final_order.state),
            date_closed=final_order.closed_at,
        )

    def get_payment_info_url(self, payment: Payment, /) -> str | None:
        """Get the URL to a payment info page."""
        return (
            "https://app.squareup.com/dashboard/sales/transactions"
            f"/{payment.external_id}"
        )

    async def update_payment(self, request: UpdatePaymentRequest, /) -> PaymentResult:
        """Update a payment."""
        source_id = request.body.get("source_id")
        if not source_id:
            raise PaymentError

        verification_token = request.body.get("verification_token")

        # Cash not permitted via web
        method = request.data.get("method")
        if method == "web" and source_id.strip().upper() == "CASH":
            raise PaymentError

        order = await self._get_order(request.external_id)
        if order.state != OrderState.open:
            raise PaymentStateError("Order is already closed")

        idempotency_key = nanoid.generate(size=14)
        amount = request.data["total_price"]
        currency = request.data["currency"]
        location_id = request.data["location_id"]
        email = request.data.get("email")

        res = await self._req(
            self._payments.create_payment,
            PaymentResponse,
            _converter.unstructure(
                CreatePayment(
                    source_id=source_id,
                    idempotency_key=idempotency_key,
                    amount_money=Money(
                        amount,
                        currency,
                    ),
                    autocomplete=False,
                    delay_duration="PT30M",
                    order_id=order.id,
                    location_id=location_id,
                    buyer_email_address=email,
                    verification_token=verification_token,
                )
            ),
        )
        if isinstance(res, ErrorResponse):
            raise PaymentError(_format_error(res))
        payment = res.payment

        idempotency_key = nanoid.generate(size=14)

        res = await self._req(
            self._orders.pay_order,
            OrderResponseBody,
            order.id,
            {"idempotency_key": idempotency_key, "payment_ids": [payment.id]},
        )
        if isinstance(res, ErrorResponse):
            raise PaymentError(_format_error(res))

        final_order = res.order
        return PaymentResult(
            id=request.id,
            service=self.id,
            external_id=final_order.id,
            status=_order_status_to_payment_status(final_order.state),
            date_closed=final_order.closed_at,
        )

    async def _get_order(self, id: str) -> Order:
        res = await self._req(self._orders.retrieve_order, OrderResponseBody, id)
        if isinstance(res, ErrorResponse):
            if res.status == 404:
                raise PaymentNotFoundError
            else:
                raise PaymentMethodError(_format_error(res))

        return res.order

    async def _sync_customers(self, cart_data: CartData) -> str | None:
        by_email = {}
        for customer in _cart_to_customers(cart_data):
            if not customer.email_address:
                continue
            by_email[customer.email_address] = customer

        last_id = None

        for email, customer in by_email.items():
            cur = await self._get_customer(email)
            res = await self._update_customer(cur, customer)
            last_id = res or last_id

        return last_id

    async def _get_customer(self, email: str) -> str | None:
        body = {"query": {"filter": {"email_address": {"exact": email}}}}
        res = await self._req(self._customers.search_customers, CustomersResponse, body)
        if isinstance(res, ErrorResponse) or not res.customers:
            return None

        return res.customers[0].id

    async def _update_customer(
        self, customer_id: str | None, customer: CreateCustomerBody
    ) -> str | None:
        if customer_id:
            body = {
                "email_address": customer.email_address,
            }

            if customer.given_name:
                body["given_name"] = customer.given_name

            if customer.family_name:
                body["family_name"] = customer.family_name

            res = await self._req(
                self._customers.update_customer, CustomerResponse, customer_id, body
            )
        else:
            res = await self._req(
                self._customers.create_customer,
                CustomerResponse,
                _converter.unstructure(customer),
            )

        if isinstance(res, ErrorResponse):
            return None
        else:
            return res.customer.id

    @property
    def _orders(self) -> OrdersApi:
        return cast(OrdersApi, self._client.orders)

    @property
    def _payments(self) -> PaymentsApi:
        return cast(PaymentsApi, self._client.payments)

    @property
    def _customers(self) -> CustomersApi:
        return cast(CustomersApi, self._client.customers)

    async def _req(
        self,
        method: Callable[_P, Optional[ApiResponse]],
        t: Type[_T],
        /,
        *a: _P.args,
        **kw: _P.kwargs,
    ) -> _T | ErrorResponse:
        res = await asyncio.to_thread(method, *a, **kw)
        assert res
        if res.is_success():
            return _converter.structure(res.body, t)
        else:
            errors = _converter.structure(res.errors, list[Error])
            return ErrorResponse(res.status_code, errors)


class SquareWebPaymentMethod:
    """Square web payment method."""

    def __init__(self, service: SquareService):
        self._service = service

    async def create_payment(self, request: CreatePaymentRequest, /) -> PaymentResult:
        """Create a payment."""
        customer_id = await self._service._sync_customers(request.cart_data)

        order_body = _build_order(
            self._service.location_id,
            customer_id,
            request.pricing_result,
            self._service._item_map,
            self._service._modifier_map,
        )

        res = await self._service._req(
            self._service._orders.create_order,
            OrderResponseBody,
            {"order": _converter.unstructure(order_body)},
        )
        if isinstance(res, ErrorResponse):
            raise PaymentMethodError(_format_error(res))

        order = res.order
        currency_str = format_currency(
            request.pricing_result.currency, request.pricing_result.total_price
        )

        return PaymentResult(
            id=request.id,
            service=self._service.id,
            external_id=order.id,
            status=_order_status_to_payment_status(order.state),
            date_created=order.created_at,
            data={
                "location_id": self._service.location_id,
                "total_price": request.pricing_result.total_price,
                "currency": request.pricing_result.currency,
                "total_price_str": currency_str,
                "environment": self._service.environment,
                "email": request.email,
                "method": "web",
            },
            body={
                "application_id": self._service.application_id,
                "location_id": self._service.location_id,
                "total_price": request.pricing_result.total_price,
                "currency": request.pricing_result.currency,
                "total_price_str": currency_str,
                "environment": self._service.environment,
            },
        )


def _build_order(
    location_id: str,
    customer_id: str | None,
    pricing_result: PricingResult,
    item_map: Mapping[str, str],
    modifier_map: Mapping[str, str],
) -> CreateOrder:
    discount_map = {}
    line_items = itertools.chain.from_iterable(
        r.line_items for r in pricing_result.registrations
    )
    sq_line_items = [
        _build_line_item(
            pricing_result.currency, discount_map, li, item_map, modifier_map
        )
        for li in line_items
    ]
    return CreateOrder(
        location_id=location_id,
        line_items=sq_line_items,
        customer_id=customer_id,
        discounts=list(discount_map.values()),
        pricing_options=PricingOptions(auto_apply_discounts=False),
    )


def _build_line_item(
    currency: str,
    discount_map: dict[str, CreateOrderLineItemDiscount],
    line_item: LineItem,
    item_map: Mapping[str, str],
    modifier_map: Mapping[str, str],
) -> CreateOrderLineItem:
    modifiers = [
        _build_modifier(currency, m, modifier_map)
        for m in line_item.modifiers
        if m.amount >= 0
    ]
    discounts = [
        _build_discount(currency, discount_map, m)
        for m in line_item.modifiers
        if m.amount < 0
    ]
    return CreateOrderLineItem(
        item_type=OrderLineItemItemType.item,
        name=line_item.name,
        catalog_object_id=item_map.get(line_item.id) if line_item.id else None,
        modifiers=modifiers,
        applied_discounts=discounts,
        base_price_money=Money(
            amount=line_item.price,
            currency=currency,
        ),
    )


def _build_modifier(
    currency: str, modifier: Modifier, modifier_map: Mapping[str, str]
) -> CreateOrderLineItemModifier:
    return CreateOrderLineItemModifier(
        catalog_object_id=modifier_map.get(modifier.id) if modifier.id else None,
        name=modifier.name,
        base_price_money=Money(
            amount=modifier.amount,
            currency=currency,
        ),
    )


def _build_discount(
    currency: str,
    discount_map: dict[str, CreateOrderLineItemDiscount],
    modifier: Modifier,
) -> CreateOrderLineItemAppliedDiscount:
    uid = nanoid.generate(size=14)
    discount = CreateOrderLineItemDiscount(
        uid=uid,
        name=modifier.name,
        type=OrderLineItemDiscountType.fixed_amount,
        scope=OrderLineItemDiscountScope.line_item,
        amount_money=Money(
            amount=-modifier.amount,
            currency=currency,
        ),
        applied_money=Money(
            amount=-modifier.amount,
            currency=currency,
        ),
    )
    discount_map[uid] = discount
    return CreateOrderLineItemAppliedDiscount(discount_uid=uid)


def _cart_to_customers(cart: CartData) -> Iterator[CreateCustomerBody]:
    regs = cart.get("registrations", [])
    for reg in regs:
        new_data = reg.get("new", {})
        yield _reg_to_customer(new_data)


def _reg_to_customer(reg: Mapping[str, Any]) -> CreateCustomerBody:
    fname = reg.get("first_name")
    lname = reg.get("last_name")
    pname = reg.get("preferred_name")
    email = reg.get("email")

    name = pname or fname

    return CreateCustomerBody(given_name=name, family_name=lname, email_address=email)


def _order_status_to_payment_status(status: str) -> PaymentStatus:
    if status == OrderState.open:
        return PaymentStatus.pending
    elif status == OrderState.canceled:
        return PaymentStatus.canceled
    elif status == OrderState.completed:
        return PaymentStatus.completed
    else:
        raise PaymentStateError(f"Unsupported status: {status}")


def _format_error(error_response: ErrorResponse) -> str:
    code = error_response.errors[0].code

    if code in _error_messages:
        return _error_messages[code]
    logger.error(f"Square error: {error_response.errors[0]}")

    return "Payment failed"


_error_messages = {
    "CARD_EXPIRED": "Card expired",
    "INVALID_EXPIRATION": "Invalid card",
    "INVALID_CARD": "Invalid card",
    "GENERIC_DECLINE": "Card declined",
    "CVV_FAILURE": "Verification failed",
    "INVALID_ACCOUNT": "Invalid card",
    "INSUFFICIENT_FUNDS": "Insufficient funds",
    "INVALID_CARD_DATA": "Invalid card",
    "CARD_DECLINED": "Card declined",
    "VERIFY_CVV_FAILURE": "Verification failed",
    "CARD_DECLINED_CALL_ISSUER": "Card declined, contact issuer",
    "CARD_DECLINED_VERIFICATION_REQUIRED": "Additional verification required",
}


def make_square_payment_service(
    config: Mapping[str, Any], converter: Converter
) -> SquareService:
    """Create a Square payment service."""
    square_config = converter.structure(config, SquareConfig)
    return SquareService(
        application_id=square_config.application_id,
        access_token=square_config.access_token,
        location_id=square_config.location_id,
        sandbox=square_config.sandbox,
        item_map=square_config.item_map,
        modifier_map=square_config.modifier_map,
    )
