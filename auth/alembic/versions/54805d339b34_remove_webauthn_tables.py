"""Remove webauthn tables

Revision ID: 54805d339b34
Revises: c7f5c2daa38e
Create Date: 2024-06-29 13:44:24.587118

"""

from typing import Sequence, Union

import sqlalchemy as sa
from alembic import op

# revision identifiers, used by Alembic.
revision: str = "54805d339b34"
down_revision: Union[str, None] = "c7f5c2daa38e"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # ### commands auto generated by Alembic - please adjust! ###
    op.drop_table("webauthn_credential")
    # ### end Alembic commands ###


def downgrade() -> None:
    # ### commands auto generated by Alembic - please adjust! ###
    op.create_table(
        "webauthn_credential",
        sa.Column(
            "auth_id", sa.VARCHAR(length=14), autoincrement=False, nullable=False
        ),
        sa.Column(
            "credential_id", sa.VARCHAR(length=300), autoincrement=False, nullable=False
        ),
        sa.Column("public_key", sa.TEXT(), autoincrement=False, nullable=False),
        sa.Column("sign_count", sa.INTEGER(), autoincrement=False, nullable=False),
        sa.ForeignKeyConstraint(
            ["auth_id"], ["auth.id"], name="fk_webauthn_credential_auth_id_auth"
        ),
        sa.PrimaryKeyConstraint("auth_id", name="pk_webauthn_credential"),
        sa.UniqueConstraint(
            "credential_id", name="uq_webauthn_credential_credential_id"
        ),
    )
    # ### end Alembic commands ###
