"""Access code module."""

import secrets
from collections.abc import Sequence
from datetime import datetime

from attrs import frozen
from cattrs.preconf.orjson import make_converter
from oes.registration.orm import Base
from oes.utils.orm import JSON
from sqlalchemy import String
from sqlalchemy.orm import Mapped, mapped_column
from typing_extensions import Self

ACCESS_CODE_LENGTH = 8
"""Length of an access code."""


@frozen
class InterviewOption:
    """Interview option."""

    id: str
    title: str | None = None


@frozen
class AccessCodeOptions:
    """Access code options."""

    add_options: Sequence[InterviewOption] = ()
    change_options: Sequence[InterviewOption] = ()
    registration_ids: Sequence[str] = ()


class AccessCode(Base):
    """Access code entity."""

    __tablename__ = "access_code"

    code: Mapped[str] = mapped_column(String(ACCESS_CODE_LENGTH), primary_key=True)
    event_id: Mapped[str]
    date_created: Mapped[datetime]
    date_expires: Mapped[datetime]
    name: Mapped[str]
    used: Mapped[bool]
    options: Mapped[JSON]

    @classmethod
    def create(
        cls,
        event_id: str,
        date_expires: datetime,
        name: str,
        options: AccessCodeOptions,
    ) -> Self:
        """Create an access code."""
        code = "".join(secrets.choice(_alphabet) for _ in range(ACCESS_CODE_LENGTH))
        entity = cls(
            code=code,
            event_id=event_id,
            date_created=datetime.now().astimezone(),
            date_expires=date_expires,
            name=name,
            used=False,
            options={},
        )
        entity.set_options(options)
        return entity

    def get_options(self) -> AccessCodeOptions:
        """Get the options."""
        return _converter.structure(self.options, AccessCodeOptions)

    def set_options(self, options: AccessCodeOptions):
        """Set the options."""
        self.options = _converter.unstructure(options)


_converter = make_converter()
_alphabet = "23456789BCDFGHJKLMNPQRSTVWXYZ"
