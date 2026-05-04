import { promises as fs } from "node:fs";
import { resolve, dirname } from "node:path";
import { glob } from "glob";
import type { ExtensionAPI, Message } from "@mariozechner/pi-coding-agent";

interface RuleFile {
  patterns: string[];
  content: string;
}

interface RuleContext {
  patterns: string[];
  content: string;
}

const RULE_DIRS = [
  ".agents/rules",
  ".pi/rules",
];

let ruleCache: RuleContext[] = [];
let lastRuleLoad = 0;
const RULE_CACHE_TTL = 5000; // 5 seconds

async function loadRules(cwd: string): Promise<RuleContext[]> {
  const now = Date.now();
  if (ruleCache.length > 0 && now - lastRuleLoad < RULE_CACHE_TTL) {
    return ruleCache;
  }

  const rules: RuleContext[] = [];
  
  for (const ruleDir of RULE_DIRS) {
    try {
      const dir = resolve(cwd, ruleDir);
      const files = await glob("*.md", { cwd: dir });
      
      for (const file of files) {
        const content = await fs.readFile(resolve(dir, file), "utf8");
        const frontmatter = parseFrontmatter(content);
        const patterns = frontmatter["applies-to"] || [];
        const bodyContent = extractBody(content);
        
        rules.push({
          patterns: Array.isArray(patterns) ? patterns : [patterns],
          content: bodyContent,
        });
      }
    } catch { /* dir doesn't exist */ }
  }
  
  ruleCache = rules;
  lastRuleLoad = now;
  return rules;
}

function parseFrontmatter(content: string): Record<string, string | string[]> {
  const result: Record<string, string | string[]> = {};
  
  const match = content.match(/^---\n([\s\S]*?)\n---\n/);
  if (!match) return result;
  
  const lines = match[1].split("\n");
  let currentKey = "";
  
  for (const line of lines) {
    const keyMatch = line.match(/^(\w+(?:-\w+)*):\s*/);
    if (keyMatch) {
      currentKey = keyMatch[1];
      const value = line.slice(keyMatch[0].length).trim();
      if (value.startsWith("[") && value.endsWith("]")) {
        result[currentKey] = value.slice(1, -1).split(",").map(s => s.trim());
      } else {
        result[currentKey] = value;
      }
    }
  }
  
  return result;
}

function extractBody(content: string): string {
  return content.replace(/^---\n[\s\S]*?\n---\n/, "");
}

function matchesPattern(filePath: string, patterns: string[]): boolean {
  for (const pattern of patterns) {
    if (pattern.includes("*")) {
      if (fnmatch(filePath, pattern)) return true;
    } else {
      if (filePath.includes(pattern)) return true;
    }
  }
  return false;
}

function fnmatch(path: string, pattern: string): boolean {
  const parts = pattern.split(/[\/*]/).filter(Boolean);
  let pathParts = path.split("/").filter(Boolean);
  
  for (const part of parts) {
    if (part === "**") {
      if (part === parts[parts.length - 1]) {
        return true; // ** at end matches everything
      }
      const nextPart = parts[parts.indexOf(part) + 1];
      const idx = pathParts.findIndex(p => p === nextPart || nextPart?.startsWith("*"));
      if (idx >= 0) {
        pathParts = pathParts.slice(idx);
      }
    } else if (part === "*") {
      if (pathParts.length === 0) return false;
      pathParts = pathParts.slice(1);
    } else if (pathParts[0]?.includes(part)) {
      pathParts = pathParts.slice(1);
    } else {
      return false;
    }
  }
  
  return pathParts.length === 0;
}

function getFileFromArgs(args: Record<string, unknown>): string | null {
  return (args.file_path as string) || 
         (args.path as string) || 
         (args.target as string) ||
         (args.command as string) ||
         null;
}

export default async function (pi: ExtensionAPI) {
  // Inject rules before agent starts
  pi.on("before_agent_start", async (event, ctx) => {
    if (!event.messages) return;
    
    const rules = await loadRules(ctx.cwd);
    
    // Check if any tools were called that would benefit from context
    const toolCalls = event.messages
      .filter(m => m.role === "assistant")
      .flatMap(m => m.content)
      .filter((c): c is { type: "toolCall"; name: string; arguments: Record<string, unknown> } => 
        typeof c === "object" && c.type === "toolCall"
      );
    
    // Collect relevant files
    const relevantPatterns = new Set<string>();
    for (const tc of toolCalls) {
      const file = getFileFromArgs(tc.arguments || {});
      if (file) {
        for (const rule of rules) {
          if (matchesPattern(file, rule.patterns)) {
            for (const pattern of rule.patterns) {
              relevantPatterns.add(pattern);
            }
          }
        }
      }
    }
    
    // Build context string
    if (relevantPatterns.size > 0) {
      const relevantRules = rules.filter(r => 
        r.patterns.some(p => relevantPatterns.has(p))
      );
      
      if (relevantRules.length > 0) {
        const context = relevantRules.map(r => r.content).join("\n\n---\n\n");
        
        // Prepend to system message or inject as context
        for (const msg of event.messages) {
          if (msg.role === "system") {
            msg.content += "\n\n## Relevant Project Rules\n\n" + context;
            break;
          }
        }
      }
    }
  });

  // Alternative: inject on context event (before each LLM call)
  pi.on("context", async (event, ctx) => {
    // Similar logic but called more frequently
    // Could be too noisy - keeping for reference
  });
}