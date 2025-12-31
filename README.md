# Redis Job Queue in Go

A lightweight background job processing system inspired by Celery and BullMQ.

## Features
- Atomic job pickup using Redis
- Acknowledgement via processing queue
- Retry with exponential backoff
- Non-blocking retries using Redis ZSET
- Dead Letter Queue (DLQ)
- Concurrent workers using goroutines

## Architecture

Producer → job_queue (LIST)
Worker → processing_queue (LIST)
Failures → retry_zset (ZSET)
Permanent failures → dlq (LIST)

## Why ZSET?
ZSET allows delayed retries without blocking workers. Jobs are scheduled using timestamps instead of sleep.

## Use cases
- Email sending
- Notifications
- Payment retries
- Background processing

## Inspired by
- Celery
- BullMQ
- Sidekiq
