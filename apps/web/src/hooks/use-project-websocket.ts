import { windowId } from "@opus/libs";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { getApiUrl } from "@/fetchers/get-api-url";
import { authClient } from "@/lib/auth-client";

function getWsUrl(path: string, params: Record<string, string> = {}) {
  const query = new URLSearchParams({ ...params, windowId });
  return `${getApiUrl(path).replace(/^http/, "ws")}?${query}`;
}

const MAX_RETRIES = 5;
const BASE_DELAY = 1000; // 1 second

type Message = Record<string, string>;

function useChannel(url: string, onMessage: (message: Message) => void) {
  const { data: session } = authClient.useSession();
  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  useEffect(() => {
    if (!url || !session?.user?.id) return;

    retriesRef.current = 0;

    function connect() {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        retriesRef.current = 0; // Reset retries on successful connection
      };

      ws.onmessage = (event) => {
        try {
          onMessageRef.current(JSON.parse(event.data));
        } catch {
          // Ignore malformed messages
        }
      };

      ws.onclose = () => {
        wsRef.current = null;

        if (retriesRef.current < MAX_RETRIES) {
          const delay = BASE_DELAY * 2 ** retriesRef.current; // 1s, 2s, 4s, 8s, 16s
          retriesRef.current += 1;
          timeoutRef.current = setTimeout(connect, delay);
        }
      };
    }
    connect();

    return () => {
      retriesRef.current = MAX_RETRIES; // Prevent reconnect after unmount
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      wsRef.current?.close();
    };
  }, [url, session?.user?.id]);
}

export function useProjectWebSocket(projectId: string) {
  const queryClient = useQueryClient();

  useChannel(
    projectId ? getWsUrl(`/ws/${encodeURIComponent(projectId)}`) : "",
    (message) => {
      if (
        message.type === "TASK_UPDATED" ||
        message.type === "TASK_CREATED" ||
        message.type === "TASK_DELETED" ||
        message.type === "TASK_LABEL_UPDATED" ||
        message.type === "TASK_MOVED" ||
        message.type === "TASK_RELATION_UPDATED" ||
        message.type === "COMMENT_UPDATED"
      ) {
        queryClient.invalidateQueries({
          queryKey: ["tasks", message.projectId],
        });

        if (message.type === "TASK_RELATION_UPDATED") {
          if (message.sourceTaskId) {
            queryClient.invalidateQueries({
              queryKey: ["task", message.sourceTaskId],
            });
            queryClient.invalidateQueries({
              queryKey: ["task-relations", message.sourceTaskId],
            });
          }
          if (message.targetTaskId) {
            queryClient.invalidateQueries({
              queryKey: ["task", message.targetTaskId],
            });
            queryClient.invalidateQueries({
              queryKey: ["task-relations", message.targetTaskId],
            });
          }
          if (!message.sourceTaskId && !message.targetTaskId) {
            queryClient.invalidateQueries({
              queryKey: ["task-relations"],
            });
          }
        } else {
          queryClient.invalidateQueries({
            queryKey: ["task", message.taskId],
          });
        }

        if (message.type === "TASK_LABEL_UPDATED") {
          queryClient.invalidateQueries({
            queryKey: ["labels", message.taskId],
          });
        }

        if (message.type === "COMMENT_UPDATED") {
          queryClient.invalidateQueries({
            queryKey: ["activities", message.taskId],
          });
          queryClient.invalidateQueries({
            queryKey: ["comments", message.taskId],
          });
        }
      }
    },
  );
}

export function useWorkspaceWebSocket(workspaceId: string) {
  const queryClient = useQueryClient();

  useChannel(workspaceId ? getWsUrl("/ws", { workspaceId }) : "", (message) => {
    if (
      message.type?.startsWith("TASK_") ||
      message.type === "COMMENT_UPDATED"
    ) {
      queryClient.invalidateQueries({ queryKey: ["my-tasks", workspaceId] });
    }
  });
}
