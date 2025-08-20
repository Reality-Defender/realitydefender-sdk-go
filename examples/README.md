
# Reality Defender Go SDK Examples

This directory contains example programs demonstrating how to use the Reality Defender Go SDK for deepfake detection and media manipulation analysis.

## Prerequisites

Before running these examples, make sure you have:

1. **Go 1.23 or later** installed
2. **Just command runner** installed: https://github.com/casey/just
3. **API Key** set as an environment variable:
   ```bash
   export REALITY_DEFENDER_API_KEY="your-api-key"
   ```
4. **Test image** placed in the `examples/images/` directory as `test_image.jpg`, or an **URL** to media stored in a social media platform.

## Running Examples

All examples can be run using the Justfile commands from the root directory:

### Basic Example
**Command:** `just run-basic`
**File:** `basic/main.go`

Demonstrates the fundamental SDK usage with:
- File upload to Reality Defender API
- Synchronous polling for detection results
- Progress indication during analysis
- Formatted output of detection scores

### Event-Based Example
**Command:** `just run-events`
**File:** `events/main.go`

Demonstrates asynchronous, event-driven detection using:
- File upload
- Event-based result handling with `client.On()` callbacks
- Non-blocking polling with `PollForResults()`
- Proper event synchronization

### Channel-Based Example
**Command:** `just run-channels`
**File:** `channels/main.go`

Shows Go-idiomatic concurrent programming using:
- Goroutines for background polling
- Channels for result communication
- Context cancellation support
- Exponential backoff for polling

### Social media analysis Example
**Command:** `just run-social <URL>`
**File:** `social/main.go`

Shows how to analyze social media content.

### Retrieve results Example
**Command:** `just run-results`
**File:** `results/main.go`

Shows how previous results can be retrieved with pagination.
