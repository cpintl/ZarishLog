"use client";

import { useState } from "react";

interface TestResult {
  name: string;
  status: "pass" | "fail" | "skipped";
  duration?: string;
  details?: string;
}

export default function TestResultsPage() {
  const [results, setResults] = useState<TestResult[] | null>(null);
  const [running, setRunning] = useState(false);
  const [output, setOutput] = useState("");

  const runTests = async (suite: string) => {
    setRunning(true);
    setOutput(`Running ${suite} tests...\n`);
    setResults(null);

    try {
      const res = await fetch(`/api/run-tests?suite=${suite}`);
      const data = await res.json();
      setResults(data.results);
      setOutput((prev) => prev + data.output);
    } catch {
      setOutput((prev) => prev + "Error: Could not run tests\n");
    } finally {
      setRunning(false);
    }
  };

  const testSuites = [
    { id: "go", name: "Go Tests", description: "FEFO, AMC, API handlers, repository" },
    { id: "go-short", name: "Go Tests (Quick)", description: "Skip slow integration tests" },
    { id: "frontend", name: "Frontend Tests", description: "React components, utilities" },
    { id: "validate", name: "Config Validation", description: "CSV, JSON, environment checks" },
    { id: "integration", name: "Integration Tests", description: "End-to-end API smoke tests" },
  ];

  const statusIcon = (status: string) => {
    switch (status) {
      case "pass": return <span className="text-green-500">✓</span>;
      case "fail": return <span className="text-red-500">✗</span>;
      case "skipped": return <span className="text-slate-400">−</span>;
      default: return <span className="text-slate-400">?</span>;
    }
  };

  return (
    <main className="mx-auto max-w-4xl px-6 py-12">
      <h1 className="text-2xl font-semibold">Test Results Dashboard</h1>
      <p className="mt-1 text-sm text-slate-500">Visual test runner for ZarishLog development sandbox</p>

      {/* Test Suite Buttons */}
      <div className="mt-8 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {testSuites.map((suite) => (
          <button
            key={suite.id}
            onClick={() => runTests(suite.id)}
            disabled={running}
            className="rounded-lg border border-slate-200 bg-white p-4 text-left hover:border-blue-300 hover:shadow-sm transition-all disabled:opacity-50"
          >
            <h3 className="font-medium text-slate-900">{suite.name}</h3>
            <p className="mt-1 text-sm text-slate-500">{suite.description}</p>
          </button>
        ))}
      </div>

      {/* Output */}
      {output && (
        <div className="mt-8">
          <h2 className="font-medium">Output</h2>
          <pre className="mt-2 max-h-96 overflow-auto rounded-lg bg-slate-900 p-4 text-sm text-green-400 font-mono">
            {output}
          </pre>
        </div>
      )}

      {/* Results Table */}
      {results && (
        <div className="mt-8">
          <h2 className="font-medium">Results</h2>
          <div className="mt-2 overflow-hidden rounded-lg border border-slate-200">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-100 text-slate-600">
                <tr>
                  <th className="px-4 py-2 w-8">Status</th>
                  <th className="px-4 py-2">Test</th>
                  <th className="px-4 py-2">Duration</th>
                  <th className="px-4 py-2">Details</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r, i) => (
                  <tr key={i} className="border-t border-slate-100">
                    <td className="px-4 py-2">{statusIcon(r.status)}</td>
                    <td className="px-4 py-2 font-medium">{r.name}</td>
                    <td className="px-4 py-2 text-slate-500">{r.duration ?? "—"}</td>
                    <td className="px-4 py-2 text-xs text-slate-400">{r.details ?? ""}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Quick Links */}
      <div className="mt-12 rounded-lg border border-slate-200 p-6">
        <h2 className="font-medium">Quick Terminal Commands</h2>
        <div className="mt-4 space-y-2">
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> make test          <span className="text-slate-500"># Run all tests</span>
          </div>
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> make test-coverage <span className="text-slate-500"># With coverage report</span>
          </div>
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> make lint         <span className="text-slate-500"># Run linters</span>
          </div>
        </div>
      </div>
    </main>
  );
}
