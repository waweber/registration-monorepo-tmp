"""Registration module."""

import uuid
from collections.abc import Callable, Mapping, Sequence
from datetime import datetime
from enum import Enum
from typing import Any, Union

from attr import fields
from attrs import define, field
from cattrs import Converter, override
from cattrs.gen import make_dict_structure_fn, make_dict_unstructure_fn
from sqlalchemy import select
from sqlalchemy.orm import Mapped, mapped_column

from oes.registration.orm import JSON, Base, Repo, UUIDStr


class Status(str, Enum):
    """Registration status values."""

    pending = "pending"
    created = "created"
    canceled = "canceled"


class StatusError(ValueError):
    """Raised when a registration's status is not acceptable."""

    pass


class Registration(Base, kw_only=True):
    """Registration entity."""

    __tablename__ = "registration"

    id: Mapped[UUIDStr] = mapped_column(
        default_factory=lambda: str(uuid.uuid4()), primary_key=True
    )
    event_id: Mapped[str]
    version: Mapped[int] = mapped_column(default=1)
    status: Mapped[Status] = mapped_column(default=Status.pending)
    date_created: Mapped[datetime] = mapped_column(
        default_factory=lambda: datetime.now().astimezone()
    )
    date_updated: Mapped[datetime | None] = mapped_column(
        default=None, onupdate=lambda: datetime.now().astimezone()
    )
    first_name: Mapped[str | None] = mapped_column(default=None)
    last_name: Mapped[str | None] = mapped_column(default=None)
    email: Mapped[str | None] = mapped_column(default=None)
    extra_data: Mapped[JSON] = mapped_column(default_factory=dict)

    __mapper_args__ = {"version_id_col": version}

    def complete(self) -> bool:
        """Complete the registration."""
        if self.status == Status.pending:
            self.status = Status.created
            return True
        elif self.status == Status.canceled:
            raise StatusError("Registration is already canceled")
        else:
            return False

    def cancel(self) -> bool:
        """Cancel the registration."""
        if self.status != Status.canceled:
            self.status = Status.canceled
            return True
        else:
            return False


@define
class RegistrationCreate:
    """Registration fields settable on creation."""

    status: Status = Status.pending
    first_name: str | None = None
    last_name: str | None = None
    email: str | None = None
    extra_data: JSON = field(factory=dict)

    def create(self, event_id: str) -> Registration:
        """Create a :class:`Registration`."""
        kw = {attr.name: getattr(self, attr.name) for attr in fields(type(self))}
        return Registration(event_id=event_id, **kw)


@define
class RegistrationUpdate:
    """Updatable registration fields."""

    first_name: str | None = None
    last_name: str | None = None
    email: str | None = None
    extra_data: JSON = field(factory=dict)

    def apply(self, registration: Registration):
        """Apply the attributes to a :class:`Registration`."""
        for attr in fields(type(self)):
            setattr(registration, attr.name, getattr(self, attr.name))


_field_names = (
    "id",
    "event_id",
    "version",
    "status",
    "date_created",
    "date_updated",
    "first_name",
    "last_name",
    "email",
)


class RegistrationRepo(Repo[Registration, Union[str, uuid.UUID]]):
    """Registration repository."""

    entity_type = Registration

    async def get(
        self,
        id: Union[str, uuid.UUID],
        *,
        event_id: str | None = None,
        lock: bool = False,
    ) -> Registration | None:
        res = await super().get(id, lock=lock)
        return None if res is None or event_id and res.event_id != event_id else res

    async def search(self, event_id: str) -> Sequence[Registration]:
        """Search registrations."""
        q = select(Registration).where(Registration.event_id == event_id)
        res = await self.session.execute(q)
        return res.scalars().all()


def make_registration_structure_fn(
    converter: Converter,
) -> Callable[[Any, Any], Registration]:
    """Make a function to structure a :class:`Registration`."""
    dict_fn = make_dict_structure_fn(
        Registration,
        converter,
    )

    def structure(v, t):
        if not isinstance(v, Mapping):
            raise ValueError(f"Invalid registration: {v}")
        extra = {k: v for k, v in v.items() if k not in _field_names}
        return dict_fn({**v, "extra_data": extra}, t)

    return structure


def make_registration_create_structure_fn(
    converter: Converter,
) -> Callable[[Any, Any], RegistrationCreate]:
    """Make a function to structure a :class:`RegistrationCreate`."""
    dict_fn = make_dict_structure_fn(
        RegistrationCreate,
        converter,
    )

    def structure(v, t):
        if not isinstance(v, Mapping):
            raise ValueError(f"Invalid registration: {v}")
        extra = {k: v for k, v in v.items() if k not in _field_names}
        reg = dict_fn({**v, "extra_data": extra}, t)
        return reg

    return structure


def make_registration_update_structure_fn(
    converter: Converter,
) -> Callable[[Any, Any], RegistrationUpdate]:
    """Make a function to structure a :class:`RegistrationUpdate`."""
    dict_fn = make_dict_structure_fn(
        RegistrationUpdate,
        converter,
    )

    def structure(v, t):
        if not isinstance(v, Mapping):
            raise ValueError(f"Invalid registration: {v}")
        extra = {k: v for k, v in v.items() if k not in _field_names}
        reg = dict_fn({**v, "extra_data": extra}, t)
        return reg

    return structure


def make_registration_unstructure_fn(
    converter: Converter,
) -> Callable[[Registration], Any]:
    """Make a function to unstructure a :class:`Registration`."""
    dict_fn = make_dict_unstructure_fn(
        Registration,
        converter,
        _cattrs_omit_if_default=True,
        version=override(omit_if_default=False),
        status=override(omit_if_default=False),
    )

    def unstructure(v: Registration) -> dict[str, Any]:
        data = dict_fn(v)
        extra = data.pop("extra_data", {})
        return {**extra, **data}

    return unstructure
