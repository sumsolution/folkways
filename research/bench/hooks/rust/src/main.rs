// Bench prototype (not a production hook). Spec: ../README.md#the-bench-hook
use regex::Regex;
use serde_json::{json, Value};
use std::io::Read;
use yaml_rust2::YamlLoader;

const COMMIT: &str = r#"\bgit\s+commit\b[^|;&]*?\s-m\s*(?:"([^"]*)"|'([^']*)')"#;
const HEADER: &str = r"^(?:feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\(([a-z0-9._-]+)\))?!?: \S";

fn deny(reason: String) {
    let out = json!({ "hookSpecificOutput": {
        "hookEventName": "PreToolUse",
        "permissionDecision": "deny",
        "permissionDecisionReason": reason,
    }});
    println!("{out}");
}

fn run() -> Result<(), String> {
    let mut input = String::new();
    std::io::stdin().read_to_string(&mut input).map_err(|e| e.to_string())?;
    let payload: Value = serde_json::from_str(&input).map_err(|e| e.to_string())?;
    if payload["tool_name"] != "Bash" {
        return Ok(());
    }
    let command = payload["tool_input"]["command"].as_str().unwrap_or("");
    let Some(m) = Regex::new(COMMIT).unwrap().captures(command) else { return Ok(()) };
    let raw = std::fs::read_to_string("team.yaml").map_err(|e| format!("team.yaml: {e}"))?;
    let docs = YamlLoader::load_from_str(&raw).map_err(|e| format!("team.yaml: {e}"))?;
    let scopes: Vec<&str> = docs
        .first()
        .and_then(|d| d["git"]["scopes"].as_vec())
        .map(|v| v.iter().filter_map(|s| s.as_str()).collect())
        .unwrap_or_default();
    let message = m.get(1).or(m.get(2)).map_or("", |g| g.as_str());
    let subject = message.split('\n').next().unwrap_or("");
    match Regex::new(HEADER).unwrap().captures(subject) {
        None => deny(format!("Commit message is not Angular \"type(scope): subject\": \"{subject}\".")),
        Some(h) => {
            let scope = h.get(1).map_or("", |g| g.as_str());
            if !scopes.is_empty() && !scopes.contains(&scope) {
                deny(format!("Scope \"{scope}\" is not in team.yaml git.scopes ({}).", scopes.join(", ")));
            }
        }
    }
    Ok(())
}

fn main() {
    if let Err(e) = run() {
        eprintln!("folkways: {e}");
        std::process::exit(1);
    }
}
