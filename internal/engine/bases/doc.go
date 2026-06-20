// Package bases provides reusable base-engine factories (xpath, json_engine,
// mediawiki, opensearch, command) and shared extraction utilities ported
// from SearXNG's engine framework.
//
// Base engines implement the engine.Engine interface and can be instantiated
// from config alone — no per-engine Go code required for simple cases.
package bases
