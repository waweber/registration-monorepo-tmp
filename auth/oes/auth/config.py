"""Config."""

import typed_settings as ts
from oes.utils.config import get_loaders
from sqlalchemy import URL, make_url


@ts.settings
class Config:
    """Config object."""

    token_secret: ts.SecretStr = ts.secret(help="the secret for signing tokens")
    db_url: URL = ts.option(
        default=make_url("postgresql+asyncpg:///auth"),
        converter=lambda v: make_url(v),
        help="the database URL",
    )
    amqp_url: str = ts.option(
        default="amqp://guest:guest@localhost/", help="the AMQP server URL"
    )
    disable_auth: bool = ts.option(default=False, help="disable auth")
    allowed_origins: list[str] = ts.option(factory=list, help="list of allowed origins")


def get_config() -> Config:
    """Load the config."""
    loaders = get_loaders("OES_AUTH_SERVICE_", ("auth.yml",))
    return ts.load_settings(Config, loaders)
