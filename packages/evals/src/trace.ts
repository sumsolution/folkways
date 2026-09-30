export interface ToolCall {
  name: string;
  input: unknown;
}

/** What one `claude -p --output-format stream-json` run did. */
export interface Trace {
  /** The final answer. */
  result: string;
  /** The agent's side of the session: its text and tool calls, not tool results. */
  transcript: string;
  toolCalls: ToolCall[];
  /** What the session loaded, from its init message. */
  skills: string[];
  tools: string[];
  plugins: string[];
  mcpServers: string[];
  costUsd: number;
  turns: number;
  isError: boolean;
}

type Json = Record<string, unknown>;

export function parseTrace(stream: string): Trace {
  const t: Trace = {
    result: '',
    transcript: '',
    toolCalls: [],
    skills: [],
    tools: [],
    plugins: [],
    mcpServers: [],
    costUsd: 0,
    turns: 0,
    isError: true,
  };
  const parts: string[] = [];
  for (const line of stream.split('\n')) {
    let m: Json;
    try {
      m = JSON.parse(line) as Json;
    } catch {
      continue;
    }
    if (m.type === 'system' && m.subtype === 'init') {
      t.skills = names(m.skills);
      t.tools = names(m.tools);
      t.plugins = names(m.plugins);
      t.mcpServers = names(m.mcp_servers);
    } else if (m.type === 'assistant') {
      const content = (m.message as Json | undefined)?.content;
      for (const block of Array.isArray(content) ? (content as Json[]) : []) {
        if (block.type === 'text' && typeof block.text === 'string') parts.push(block.text);
        if (block.type === 'tool_use' && typeof block.name === 'string') {
          t.toolCalls.push({ name: block.name, input: block.input });
          parts.push(`[tool ${block.name}] ${JSON.stringify(block.input)}`);
        }
      }
    } else if (m.type === 'result') {
      t.result = typeof m.result === 'string' ? m.result : '';
      t.costUsd = typeof m.total_cost_usd === 'number' ? m.total_cost_usd : 0;
      t.turns = typeof m.num_turns === 'number' ? m.num_turns : 0;
      t.isError = m.is_error === true;
    }
  }
  t.transcript = parts.join('\n');
  return t;
}

function names(list: unknown): string[] {
  if (!Array.isArray(list)) return [];
  return list.map((x) => (typeof x === 'string' ? x : String((x as Json).name ?? ''))).filter(Boolean);
}
