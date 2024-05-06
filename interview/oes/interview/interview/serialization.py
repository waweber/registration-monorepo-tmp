"""Serialization module."""

from collections.abc import Callable, Mapping
from typing import Any

from cattrs import Converter
from oes.interview.interview.step_types.ask import AskStep
from oes.interview.interview.step_types.exit import ExitStep
from oes.interview.interview.step_types.set import SetStep


def make_step_structure_fn(converter: Converter) -> Callable[[Any, Any], Any]:
    """Make a function to structure a :class:`Step`."""

    def structure(v, t):
        if isinstance(v, Mapping):
            if "ask" in v:
                return converter.structure(v, AskStep)
            elif "exit" in v:
                return converter.structure(v, ExitStep)
            elif "set" in v:
                return converter.structure(v, SetStep)
        raise ValueError(f"Invalid step: {v}")

    return structure
