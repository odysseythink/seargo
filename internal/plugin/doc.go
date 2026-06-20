// Package plugin provides a plugin system for extending SearGo behavior.
//
// Plugins can hook into the search pipeline at four points:
//   - pre_search: before engine selection (can abort search)
//   - on_result: per engine result (can filter or mutate)
//   - post_search: after results are merged (can append results)
//
// Built-in plugins are registered via init() in internal/plugin/builtin/.
// Third-party plugins are loaded from .so files via plugin.Open.
package plugin
