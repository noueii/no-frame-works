/**
 * Go Backend Harness Extension
 *
 * Enforces the no-frame-works layered architecture by:
 * - Injecting harness guidelines into system prompts
 * - Registering /harness-review command
 * - Providing code review capabilities
 *
 * Install: Copy to ~/.pi/agent/extensions/ or project .pi/extensions/
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { Type } from "@sinclair/typebox";
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

// HARNESS.md content cached at startup
let harnessContent = "";

function loadHarnessRules(cwd: string): string {
  const harnessPath = path.join(cwd, ".pi", "harness", "HARNESS.md");
  if (fs.existsSync(harnessPath)) {
    return fs.readFileSync(harnessPath, "utf-8");
  }
  return "";
}

function loadRulesDir(cwd: string): string[] {
  const rulesPath = path.join(cwd, ".agents", "rules");
  if (!fs.existsSync(rulesPath)) return [];
  
  const files = fs.readdirSync(rulesPath);
  return files.filter(f => f.endsWith(".md")).map(f => path.join(rulesPath, f));
}

export default function harnessExtension(pi: ExtensionAPI) {
  // Cache harness content on session start
  pi.on("session_start", async (_event, ctx) => {
    harnessContent = loadHarnessRules(ctx.cwd);
  });

  // Inject harness into system prompt
  pi.on("before_agent_start", async (event, ctx) => {
    const rulesFiles = loadRulesDir(ctx.cwd);
    
    // Build rules listing
    let rulesSection = "";
    if (rulesFiles.length > 0) {
      rulesSection = rulesFiles.map(f => {
        const name = path.basename(f, ".md");
        return `- \`.agents/rules/${path.basename(f)}\` — ${name}`;
      }).join("\n");
    }

    // Build harness injection
    let harnessSection = "";
    if (harnessContent) {
      // Extract the Quick Reference section for brevity
      const quickRefMatch = harnessContent.match(/## Quick Reference[\s\S]*?(?=##|$)/);
      harnessSection = quickRefMatch 
        ? `\n\n## HARNESS QUICK REFERENCE\n\n${quickRefMatch[0]}`
        : "";
    }

    const injection = `

## ARCHITECTURE RULES

Type isolation — layers must NOT cross:
- Handler: \`oapi.*\` + API contract only. MUST NOT use \`domain.*\` or \`model.*\`
- Service: API contract + \`domain.*\` only. MUST NOT use \`oapi.*\` or \`model.*\`
- Repository: \`domain.*\` + \`model.*\` only. MUST NOT use \`oapi.*\`

Error handling — use only \`github.com/go-errors/errors\`:
- NEVER \`fmt.Errorf\` or stdlib \`errors\`
- Wrap with \`%w\`: \`errors.Errorf("layer.service.op: %w", err)\`
- Log once at handler, never silently swallow errors

${rulesSection ? `## AVAILABLE RULE RUBRICS\n\n${rulesSection}\n\nUse the read tool to load specific rules when needed.` : ""}
${harnessSection}
`;
    return { systemPrompt: event.systemPrompt + injection };
  });

  // Register /harness-review command
  pi.registerCommand("harness-review", {
    description: "Review code against architecture rules",
    handler: async (args, ctx) => {
      ctx.ui.notify("Running harness review...", "info");
      
      // Get target (default to recent changes)
      const target = args || "";
      
      // Find files to review
      let filesToReview: string[] = [];
      
      if (target) {
        // User specified a target
        if (fs.existsSync(path.join(ctx.cwd, target))) {
          // It's a path, find all .go files
          filesToReview = findGoFiles(path.join(ctx.cwd, target));
        } else {
          // It's a specific file
          filesToReview = [target];
        }
      } else {
        // Get git diff
        const { stdout } = await pi.exec("git", ["diff", "--name-only", "HEAD"], { cwd: ctx.cwd });
        const { stdout: staged } = await pi.exec("git", ["diff", "--cached", "--name-only"], { cwd: ctx.cwd });
        
        const changed = [...new Set([...stdout.trim().split("\n"), ...staged.trim().split("\n")])]
          .filter(f => f.endsWith(".go") && !f.includes("vendor"));
        
        filesToReview = changed;
      }

      if (filesToReview.length === 0) {
        ctx.ui.notify("No Go files to review", "info");
        return;
      }

      ctx.ui.notify(`Reviewing ${filesToReview.length} file(s)...`, "info");

      // Read the rules for reference
      const rulesPath = path.join(ctx.cwd, ".agents", "rules");
      const ruleFiles = fs.existsSync(rulesPath) 
        ? fs.readdirSync(rulesPath).filter(f => f.endsWith(".md")).map(f => f.replace(".md", ""))
        : [];

      // Build review result
      const review = [
        `## Harness Review: ${filesToReview.length} file(s)`,
        "",
        "Reviewing against rules in `.agents/rules/`",
        ""
      ];

      // Send review as follow-up message
      pi.sendMessage({
        customType: "harness-review",
        content: review.join("\n"),
        display: true,
        details: { files: filesToReview, ruleCount: ruleFiles.length }
      }, { deliverAs: "followUp", triggerTurn: true });
    },
  });

  // Register harness-review tool
  pi.registerTool({
    name: "harness_review",
    label: "Harness Review",
    description: "Review Go code for architecture violations. Use after implementing new endpoints to verify compliance with handler/service/repository/domain layer rules.",
    promptSnippet: "Review code for architecture compliance",
    promptGuidelines: [
      "Use harness_review after implementing new endpoints to check layer isolation",
      "Use harness_review when asked to review code against project architecture rules"
    ],
    parameters: Type.Object({
      target: Type.Optional(Type.String({ description: "File or directory to review. Omit for git changes." })),
      layer: Type.Optional(Type.String({ description: "Specific layer to focus on: handler, service, repository, domain, or flow" })),
    }),
    async execute(_toolCallId, params, signal, _onUpdate, ctx) {
      const cwd = ctx.cwd;
      let filesToReview: string[] = [];
      const target = params.target || "";

      if (target) {
        const fullPath = path.join(cwd, target);
        if (fs.existsSync(fullPath)) {
          if (fs.statSync(fullPath).isDirectory()) {
            filesToReview = findGoFiles(fullPath);
          } else {
            filesToReview = [target];
          }
        } else {
          return {
            content: [{ type: "text", text: `Target not found: ${target}` }],
            details: { error: "path_not_found" }
          };
        }
      } else {
        // Get git changes
        try {
          const { stdout } = await pi.exec("git", ["diff", "--name-only", "HEAD"], { cwd, signal });
          const { stdout: staged } = await pi.exec("git", ["diff", "--cached", "--name-only"], { cwd, signal });
          
          filesToReview = [...new Set([...stdout.trim().split("\n"), ...staged.trim().split("\n")])]
            .filter(f => f.endsWith(".go") && !f.includes("vendor"));
        } catch {
          return {
            content: [{ type: "text", text: "Could not determine changed files. Specify a target." }],
            details: { error: "git_failed" }
          };
        }
      }

      if (filesToReview.length === 0) {
        return {
          content: [{ type: "text", text: "No Go files to review." }],
          details: { result: "no_files" }
        };
      }

      // Classify files by layer and run targeted checks
      const findings: string[] = [];
      
      for (const file of filesToReview) {
        const fullPath = path.join(cwd, file);
        if (!fs.existsSync(fullPath)) continue;
        
        const content = fs.readFileSync(fullPath, "utf-8");
        const layerChecks = checkLayerViolations(file, content, params.layer);
        findings.push(...layerChecks);
      }

      if (findings.length === 0) {
        return {
          content: [{ type: "text", text: `✅ Review complete: ${filesToReview.length} file(s). No violations found.` }],
          details: { files: filesToReview, violations: 0 }
        };
      }

      return {
        content: [{ type: "text", text: `⚠️ Review complete: ${filesToReview.length} file(s), ${findings.length} finding(s):\n\n${findings.join("\n")}` }],
        details: { files: filesToReview, violations: findings.length, findings }
      };
    },
  });
}

/**
 * Find all .go files in a directory recursively
 */
function findGoFiles(dir: string): string[] {
  const results: string[] = [];
  
  if (!fs.existsSync(dir)) return results;
  
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    
    if (entry.isDirectory()) {
      // Skip vendor and test directories
      if (entry.name === "vendor" || entry.name === "testdata") continue;
      results.push(...findGoFiles(fullPath));
    } else if (entry.name.endsWith(".go")) {
      results.push(fullPath);
    }
  }
  
  return results;
}

/**
 * Check a file for layer-specific violations
 */
function checkLayerViolations(file: string, content: string, layerFilter?: string): string[] {
  const findings: string[] = [];
  const lines = content.split("\n");

  // Determine which layer this file is in
  const isHandler = file.includes("/handler/") || file.includes("/webserver/handler/");
  const isService = file.includes("/services/") || file.includes("/service/");
  const isRepository = file.includes("/repository/");
  const isDomain = file.includes("/domain/");

  const layer = isHandler ? "handler" : isService ? "service" : isRepository ? "repository" : isDomain ? "domain" : "unknown";

  // Apply filter
  if (layerFilter && layer !== layerFilter) return findings;

  // Common violations (all layers)
  checkCommonViolations(file, content, lines, findings);

  // Layer-specific checks
  if (layer === "handler" || layerFilter === "handler") {
    checkHandlerViolations(file, content, lines, findings);
  }
  
  if (layer === "service" || layerFilter === "service") {
    checkServiceViolations(file, content, lines, findings);
  }
  
  if (layer === "repository" || layerFilter === "repository") {
    checkRepositoryViolations(file, content, lines, findings);
  }
  
  if (layer === "domain" || layerFilter === "domain") {
    checkDomainViolations(file, content, lines, findings);
  }

  return findings;
}

function checkCommonViolations(file: string, content: string, lines: string[], findings: string[]) {
  // Check for fmt.Errorf (should use go-errors)
  const fmtErrorfRegex = /fmt\.Errorf\s*\(/g;
  let match;
  while ((match = fmtErrorfRegex.exec(content)) !== null) {
    const lineNum = content.substring(0, match.index).split("\n").length;
    findings.push(`🟠 conv: ${file}:L${lineNum} — fmt.Errorf found. Use errors.Errorf from go-errors instead.`);
  }

  // Check for stdlib errors import
  if (content.includes('import "errors"') && !content.includes('"github.com/go-errors/errors"')) {
    findings.push(`🟠 conv: ${file} — stdlib "errors" imported. Use "github.com/go-errors/errors" instead.`);
  }
}

function checkHandlerViolations(file: string, content: string, lines: string[], findings: string[]) {
  // Check for domain.* import
  if (content.includes('"domain/') || content.includes("domain.")) {
    findings.push(`🟠 conv: ${file} — domain import in handler. Handlers must NOT use domain.*.`);
  }

  // Check for model.* import  
  if (content.includes('"model/') || content.includes("model.")) {
    findings.push(`🟠 conv: ${file} — model import in handler. Handlers must NOT use model.*.`);
  }

  // Check for direct repo access (should go through API)
  if (content.includes("repo.") && !content.includes("post.PostRepository")) {
    findings.push(`🟠 conv: ${file} — possible direct repo call. Use module API interface instead.`);
  }
}

function checkServiceViolations(file: string, content: string, lines: string[], findings: string[]) {
  // Check for oapi.* import
  if (content.includes('"oapi') || content.includes("oapi.")) {
    findings.push(`🟠 conv: ${file} — oapi import in service. Services must NOT use oapi.*.`);
  }

  // Check for model.* import
  if (content.includes('"model/') || content.includes("model.")) {
    findings.push(`🟠 conv: ${file} — model import in service. Services must NOT use model.*.`);
  }

  // Check for Validate() call
  if (content.includes("func ") && content.includes("Request") && !content.includes("Validate()")) {
    // Check if this is a service function (not a method definition)
    const linesWithFunc = lines.filter(l => l.includes("func (") && l.includes("Request"));
    if (linesWithFunc.length > 0 && !content.includes("req.Validate()")) {
      findings.push(`🟠 conv: ${file} — service function missing req.Validate() call.`);
    }
  }
}

function checkRepositoryViolations(file: string, content: string, lines: string[], findings: string[]) {
  // Check for oapi import
  if (content.includes('"oapi') || content.includes("oapi.")) {
    findings.push(`🟠 conv: ${file} — oapi import in repository. Repos must NOT use oapi.*.`);
  }

  // Check for raw SQL
  const rawSqlPatterns = [
    /db\.Query/,
    /db\.QueryRow/,
    /db\.Exec/,
    /\.QueryContext/,
    /\.ExecContext/,
  ];
  
  for (const pattern of rawSqlPatterns) {
    if (pattern.test(content)) {
      findings.push(`🟠 conv: ${file} — raw SQL pattern found. Use go-jet query builder exclusively.`);
      break;
    }
  }

  // Check for .SET() in update (should use MutableColumns)
  if (content.includes(".UPDATE(") && content.includes(".SET(")) {
    findings.push(`🟡 risk: ${file} — manual column update. Prefer MODEL() with MutableColumns.`);
  }
}

function checkDomainViolations(file: string, content: string, lines: string[], findings: string[]) {
  // Check for infrastructure imports
  const infraPatterns = [
    /"database\/sql"/,
    /"net\/http"/,
    /"github\.com\/lib\/pq"/,
    /"github\.com\/go-chi\//,
  ];

  for (const pattern of infraPatterns) {
    if (pattern.test(content)) {
      findings.push(`🟠 conv: ${file} — infrastructure import in domain. Domain must be pure.`);
      break;
    }
  }

  // Check for business logic violations (methods doing I/O)
  // This is harder to detect, so we just flag obvious patterns
  if (content.includes("db.") || content.includes("http.") || content.includes("client.")) {
    findings.push(`🟠 conv: ${file} — possible I/O in domain. Domain should be pure business logic.`);
  }
}