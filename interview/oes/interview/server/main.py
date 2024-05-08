"""Main entry points."""

import os
import sys

from cattrs.preconf.orjson import make_converter
from oes.interview.config.config import get_config, load_config_file
from oes.interview.serialization import configure_converter
from oes.interview.server.routes import response_converter, routes
from oes.interview.storage import StorageService
from oes.utils.sanic import setup_app
from sanic import Sanic


def main():
    """CLI wrapper."""
    os.execlp(
        "sanic",
        "oes-interview-service",
        "oes.interview.server.main:create_app",
        *sys.argv[1:]
    )


def create_app():
    """Main app."""
    app = Sanic("Interview")
    config = get_config()

    converter = make_converter()
    configure_converter(converter)
    configure_converter(response_converter.converter)

    interview_config = load_config_file(config.config_file, converter)
    interviews = interview_config.get_interviews(
        config.config_file.resolve().parent, converter
    )

    setup_app(app, converter=converter)
    app.ext.dependency(config)
    app.ctx.interviews = interviews

    app.blueprint(routes)

    @app.before_server_start
    async def setup_redis(app: Sanic):
        storage = StorageService(config.redis_url, converter)
        app.ctx.storage = storage
        app.ext.dependency(storage)

    @app.after_server_stop
    async def stop_redis(app: Sanic):
        await app.ctx.storage.aclose()

    return app
