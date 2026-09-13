import { useCallback, useEffect, useState } from "react";
import { fetchFile, fetchGraph, saveFile } from "@/lib/config-studio/api";
import type {
  FilePayload,
  FileUpdateRequest,
  GraphPayload,
  StudioDomain,
} from "@/lib/config-studio/types";

export interface Notice {
  type: "success" | "error";
  text: string;
}

export function useConfigStudio() {
  const [domain, setDomain] = useState<StudioDomain>("overview");
  const [payload, setPayload] = useState<GraphPayload | null>(null);
  const [payloadLoading, setPayloadLoading] = useState(true);
  const [payloadError, setPayloadError] = useState<string | null>(null);

  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [filePath, setFilePath] = useState<string | null>(null);
  const [file, setFile] = useState<FilePayload | null>(null);
  const [fileLoading, setFileLoading] = useState(false);
  const [fileError, setFileError] = useState<string | null>(null);
  const [edit, setEdit] = useState<FileUpdateRequest | null>(null);
  const [saving, setSaving] = useState(false);
  const [savedAt, setSavedAt] = useState<string | null>(null);
  const [notice, setNotice] = useState<Notice | null>(null);

  const flashNotice = useCallback((type: Notice["type"], text: string) => {
    setNotice({ type, text });
    window.setTimeout(() => setNotice((n) => (n?.text === text ? null : n)), 4000);
  }, []);

  const loadGraph = useCallback(async (d: StudioDomain) => {
    const next = await fetchGraph(d);
    setPayload(next);
    setPayloadError(null);
    setPayloadLoading(false);
  }, []);

  const refresh = useCallback(async () => {
    setPayloadLoading(true);
    try {
      await loadGraph(domain);
    } catch (err) {
      setPayloadError(err instanceof Error ? err.message : "Failed to load graph");
      setPayloadLoading(false);
    }
  }, [domain, loadGraph]);

  useEffect(() => {
    let cancelled = false;
    void fetchGraph(domain)
      .then((next) => {
        if (cancelled) return;
        setPayload(next);
        setPayloadError(null);
        setPayloadLoading(false);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setPayloadError(err instanceof Error ? err.message : "Failed to load graph");
        setPayloadLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [domain]);

  const loadFile = useCallback(async (path: string) => {
    setFileLoading(true);
    setFileError(null);
    setFilePath(path);
    try {
      setFile(await fetchFile(path));
      setEdit(null);
    } catch (err) {
      setFileError(err instanceof Error ? err.message : "Failed to load file");
      setFile(null);
    } finally {
      setFileLoading(false);
    }
  }, []);

  const select = useCallback(
    (id: string | null, refs: string[] = []) => {
      setSelectedId(id);
      setEdit(null);
      const ref = refs.find((r) => r.startsWith("config/") || r.startsWith("docs/"));
      if (ref) {
        void loadFile(ref);
      } else {
        setFilePath(null);
        setFile(null);
        setFileError(null);
      }
    },
    [loadFile],
  );

  const closePanel = useCallback(() => {
    setSelectedId(null);
    setFilePath(null);
    setFile(null);
    setEdit(null);
    setFileError(null);
  }, []);

  const save = useCallback(async () => {
    if (!edit || !filePath) return;
    setSaving(true);
    try {
      const result = await saveFile(filePath, edit);
      setFile(result.payload);
      setEdit(null);
      setSavedAt(new Date().toLocaleTimeString());
      flashNotice("success", `Saved ${filePath}`);
      setPayloadLoading(true);
      const next = await fetchGraph(domain);
      setPayload(next);
      setPayloadError(null);
      setPayloadLoading(false);
    } catch (err) {
      flashNotice("error", err instanceof Error ? err.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }, [edit, filePath, domain, flashNotice]);

  useEffect(() => {
    if (!filePath || edit || saving) return;
    const pollTimer = window.setInterval(() => {
      void fetchFile(filePath)
        .then((updated) => {
          // Only adopt external changes when the user is not mid-edit.
          setFile((current) => (current?.entry.path === filePath ? updated : current));
          setSavedAt(new Date().toLocaleTimeString());
        })
        .catch(() => {
          // silent — next poll will retry
        });
    }, 8000);
    return () => window.clearInterval(pollTimer);
  }, [filePath, edit, saving]);

  const selectedNode = payload?.nodes.find((n) => n.id === selectedId) ?? null;

  return {
    domain,
    setDomain,
    payload,
    payloadLoading,
    payloadError,
    refresh,
    selectedNode,
    select,
    closePanel,
    file,
    fileLoading,
    fileError,
    edit,
    setEdit,
    save,
    saving,
    savedAt,
    notice,
  };
}

export type UseConfigStudio = ReturnType<typeof useConfigStudio>;
