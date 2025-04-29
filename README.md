# MCP Server Scanner

## Overview

This tool detects MCP Servers running over Server-Sent Events (SSE) and enumerates its available tools.

**⚠️ IMPORTANT: This is a quick prototype and is NOT production-ready. The code is intentionally sloppy as it was built for rapid feedback and as such it may miss some edge cases..**

## Purpose

This tool was created to test hypotheses about MCP Servers exposed to the internet. With ongoing discussions around MCP Authorization, it provides a way to:

- Identify internet-exposed MCP servers
- Determine if they implement proper security measures

## Usage

Provide the base HTTP URL with an optional timeout parameter:

```
./mcp-scanner -url http://REDACTED -timeout 5                                       
2025/04/29 12:16:56 Initiating connection
2025/04/29 12:16:56 Processing SSE stream
2025/04/29 12:16:56 Waiting for messages
2025/04/29 12:16:56 Sending message: /messages?sessionId=4ac0f3be-619e-4fc2-8b5b-cda90df2392c
2025/04/29 12:16:56 Sent message: /messages?sessionId=4ac0f3be-619e-4fc2-8b5b-cda90df2392c
2025/04/29 12:16:56 Received message: /messages?sessionId=4ac0f3be-619e-4fc2-8b5b-cda90df2392c
2025/04/29 12:16:56 Found messages endpoint: messages?sessionId=4ac0f3be-619e-4fc2-8b5b-cda90df2392c
2025/04/29 12:16:56 Successfully got the messagesEndpoint: messages?sessionId=4ac0f3be-619e-4fc2-8b5b-cda90df2392c
2025/04/29 12:16:56 Sending message: {"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{},"resources":{}},"serverInfo":{"name":"mcp-sse-server","version":"1.0.0"}},"jsonrpc":"2.0","id":0}
2025/04/29 12:16:56 Sent message: {"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{},"resources":{}},"serverInfo":{"name":"mcp-sse-server","version":"1.0.0"}},"jsonrpc":"2.0","id":0}
2025/04/29 12:16:56 Received message: {"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{},"resources":{}},"serverInfo":{"name":"mcp-sse-server","version":"1.0.0"}},"jsonrpc":"2.0","id":0}
2025/04/29 12:16:56 Sending message: {"result":{"tools":[{"name":"add","inputSchema":{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}},"required":["a","b"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}},{"name":"search","inputSchema":{"type":"object","properties":{"query":{"type":"string"},"count":{"type":"number"}},"required":["query"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}}]},"jsonrpc":"2.0","id":1}
2025/04/29 12:16:56 Sent message: {"result":{"tools":[{"name":"add","inputSchema":{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}},"required":["a","b"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}},{"name":"search","inputSchema":{"type":"object","properties":{"query":{"type":"string"},"count":{"type":"number"}},"required":["query"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}}]},"jsonrpc":"2.0","id":1}
2025/04/29 12:16:56 Received message: {"result":{"tools":[{"name":"add","inputSchema":{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}},"required":["a","b"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}},{"name":"search","inputSchema":{"type":"object","properties":{"query":{"type":"string"},"count":{"type":"number"}},"required":["query"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}}]},"jsonrpc":"2.0","id":1}
2025/04/29 12:16:56 Tools: {"result":{"tools":[{"name":"add","inputSchema":{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}},"required":["a","b"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}},{"name":"search","inputSchema":{"type":"object","properties":{"query":{"type":"string"},"count":{"type":"number"}},"required":["query"],"additionalProperties":false,"$schema":"http://json-schema.org/draft-07/schema#"}}]},"jsonrpc":"2.0","id":1}
```
