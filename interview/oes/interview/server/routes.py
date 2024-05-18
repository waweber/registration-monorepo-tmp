"""Routes."""

from collections.abc import Mapping
from typing import Any, cast

from attrs import field, frozen
from immutabledict import immutabledict
from oes.interview.input.question import ValidationError
from oes.interview.interview.interview import Interview, make_interview_context
from oes.interview.interview.state import InterviewState
from oes.interview.interview.step_types.ask import AskResult
from oes.interview.interview.step_types.exit import ExitResult
from oes.interview.interview.update import update_interview
from oes.interview.storage import StorageService
from oes.utils.request import BadRequest, CattrsBody, raise_not_found
from oes.utils.response import ResponseConverter
from sanic import Blueprint, Request

routes = Blueprint("interview")


def _json_default(v):
    if isinstance(v, immutabledict):
        return dict(v)
    else:
        raise TypeError(type(v))


response_converter = ResponseConverter(json_default=_json_default)


@frozen
class InterviewStartRequest:
    """Request body to start an interview."""

    target: str | None = None
    context: Mapping[str, Any] = field(factory=dict)
    data: Mapping[str, Any] = field(factory=dict)


@frozen
class InterviewUpdateRequest:
    """Request body to start an interview."""

    state: str
    responses: Mapping[str, Any] | None = None


@frozen
class InterviewResponse:
    """Interview response body."""

    state: str
    completed: bool
    content: AskResult | ExitResult | None


@routes.post("/interviews/<interview_id>")
@response_converter
async def start_interview(
    request: Request, storage: StorageService, interview_id: str, body: CattrsBody
) -> InterviewResponse:
    """Start an interview."""
    interviews: Mapping[str, Interview] = request.app.ctx.interviews
    interview = raise_not_found(interviews.get(interview_id))

    req = await body(InterviewStartRequest)

    state = InterviewState(
        target=req.target,
        context=req.context,
        data=req.data,
    )
    context = make_interview_context(interview.questions, interview.steps, state)
    key = await storage.put(context)
    return InterviewResponse(key, state.completed, None)


@routes.post("/update-interview")
@response_converter
async def update_interview_route(
    request: Request, storage: StorageService, body: CattrsBody
) -> InterviewResponse:
    """Update an interview."""
    req = await body(InterviewUpdateRequest)
    context = raise_not_found(await storage.get(req.state))

    try:
        result_ctx, content = await update_interview(context, req.responses)
    except ValidationError:
        exc = BadRequest("Invalid input")
        exc.status_code = 422
        raise exc
    key = await storage.put(result_ctx)
    return InterviewResponse(key, result_ctx.state.completed, cast(AskResult, content))