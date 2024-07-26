"""Message queue module."""

from __future__ import annotations

import asyncio
from collections.abc import Mapping
from contextlib import suppress
from typing import Any

import aio_pika
import orjson
from cattrs import Converter
from loguru import logger
from oes.payment.config import Config


class MQService:
    """Message queue service."""

    def __init__(self, config: Config, converter: Converter):
        self.config = config
        self.converter = converter
        self._ready = asyncio.Event()
        self.ready = False

    async def start(self):
        """Start the service."""
        self.run_task = asyncio.create_task(self._run())

    async def _run(self):
        conn = await aio_pika.connect_robust(
            self.config.amqp_url,
            fail_fast=False,
        )
        async with conn:
            channel = await conn.channel()
            self.exchange = await channel.declare_exchange(
                "email", type=aio_pika.ExchangeType.TOPIC, durable=True
            )
            self._ready.set()
            self.ready = True
            logger.debug("AMQP connected")
            while True:
                await asyncio.sleep(3600)

    async def publish(self, key: str, body: Mapping[str, Any]):
        """Publish a message."""
        msg = aio_pika.Message(
            orjson.dumps(body), delivery_mode=aio_pika.DeliveryMode.PERSISTENT
        )
        await self._ready.wait()
        await self.exchange.publish(msg, key)

    async def stop(self):
        """Stop the service."""
        self.run_task.cancel()
        with suppress(asyncio.CancelledError):
            await self.run_task
