# Canvas Server

Silly project that I got obssessed over for a good week (and rewrote at least 5 times) that lets people draw together in real-time. Not proud of the jank code but it works surprisingly well so might as well just leave it here.

Here's what we got:

- Real-time pixel placement and updates
- WebSocket streaming for instant feedback on the front end
- Simple HTTP API for getting/setting pixels as well as an optional websocket drawing endpoint
- Auto-saves the canvas every minute and on shutdown (mostly "just in case")
- Default 1024x1024 canvas (handles much larger ones too)
- No rate limiting (more fun, and less work)
- Drawing tested with ~2k/s via http, 50k+/s via WebSocket (eats bandwidth due to no ws compression)

## How It Works

The server handles:

- HTTP endpoints for basic interactions
- WebSocket connections for real-time drawing (optional)
- WebSocket streaming to watch the canvas update live
- Canvas state management and persistence

## Running it

Run the server and visit http://localhost:9999
(Set WS_DRAW=1 to enable WebSocket drawing - no rate limits, be careful)

## Documentation

Full API documentation is available at `/docs`

# Note

This was a fun project I got obsessed with for a week, the code doesn't look good but works well enough. It's in a pretty stable state so I didn't want to spend more time on just cleaning it up. I don't plan to keep maintaining it but if you have any questions reach out or open an issue.
