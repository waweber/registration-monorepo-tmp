"""Ask step."""

from typing import Literal

from attrs import frozen
from oes.interview.input.types import JSONSchema
from oes.interview.interview.error import InterviewError
from oes.interview.interview.interview import InterviewContext
from oes.interview.interview.update import UpdateResult
from oes.utils.logic import WhenCondition


@frozen(kw_only=True)
class AskResult:
    """Response content containing a question."""

    type: Literal["question"] = "question"
    schema: JSONSchema


@frozen
class AskStep:
    """A step that asks a question."""

    ask: str
    when: WhenCondition = True

    def __call__(self, context: InterviewContext) -> UpdateResult:
        if self.ask in context.state.answered_question_ids:
            return UpdateResult(context)
        question_template = context.question_templates.get(self.ask)
        if question_template is None:
            raise InterviewError(f"No question with ID {self.ask}")
        question = question_template.get_question(context.state.template_context)
        state = context.state.update(
            current_question=question_template,
            answered_question_ids=context.state.answered_question_ids | {self.ask},
        )
        content = AskResult(schema=question.schema)
        return UpdateResult(context.with_state(state), content)
