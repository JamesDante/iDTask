"use client";

import { useEffect, useState } from "react";
import { ReactFlow,
    Background,
    Controls,
    MiniMap,
    type Node,
    type Edge,
    type ReactFlowInstance,
  } from "@xyflow/react";
import '@xyflow/react/dist/style.css';

import { apiClient } from "@/lib/api-client"; // 之前封装的 client

interface DagResponse {
  nodes: Node[];
  edges: Edge[];
}

export default function TaskDag({ taskId }: { taskId: string }) {
  const [elements, setElements] = useState<{ nodes: Node[]; edges: Edge[] }>({
    nodes: [],
    edges: [],
  });

  useEffect(() => {
    async function fetchDag() {
      try {
        //var data = await apiClient.get<DagResponse>(`/dag/${taskId}`);
        var data = {
            "nodes": [
              {
                "id": "A",
                "type": "default",
                "data": { "label": "Start" },
                "position": { "x": 0, "y": 0 }
              },
              {
                "id": "B",
                "data": { "label": "Process" },
                "position": { "x": 200, "y": 100 }
              }
            ],
            "edges": [
              { "id": "e1-2", "source": "A", "target": "B", "animated": true }
            ]
          };
        setElements({ nodes: data.nodes, edges: data.edges });
      } catch (err) {
        console.error("Failed to fetch DAG:", err);
      }
    }
    fetchDag();
  }, [taskId]);

  const onInit = (instance: ReactFlowInstance) => instance.fitView();

  return (
    <div className="h-[600px] w-full border rounded-md">
      <ReactFlow colorMode="dark"
        nodes={elements.nodes}
        edges={elements.edges}
        onInit={onInit}
        fitView
      >
        <Background gap={12} size={1} />
        <Controls />
      </ReactFlow>
    </div>
  );
}