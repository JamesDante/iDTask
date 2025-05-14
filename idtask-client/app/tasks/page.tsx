// app/page.tsx
"use client"

import { AppSidebar } from "@/components/app-sidebar";
import { DataTable } from "@/components/data-table";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { apiClient } from "@/lib/api-client";
import { useState } from "react";

import data from "../dashboard/data.json"

export default function HomePage() {
  const [tasks, setTasks] = useState(data);
  const [newTask, setNewTask] = useState({ type: "webpage", status: "Pending" });

  const addTask = async () => {
    //if (!newTask.id) return;

    try {
      const response = await apiClient.post("/tasks", newTask)
      
      if (response.ok) {
        setTasks([...tasks, newTask]);
        setNewTask({ id: "", status: "Pending" });
      } else {
        console.error("Failed to add task");
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  return (
  <SidebarProvider
    style={
      {
        "--sidebar-width": "calc(var(--spacing) * 72)",
        "--header-height": "calc(var(--spacing) * 12)",
      } as React.CSSProperties
    }
  >
  <AppSidebar variant="inset" />
    <SidebarInset>
      <SiteHeader />
      <main className="p-6">
      <h1 className="text-2xl font-bold mb-4">Task List</h1>
      <div className="mb-4 space-x-2">
        {/* <input
          className="border px-2 py-1"
          placeholder="Task ID"
          value={newTask.id}
          onChange={(e) => setNewTask({ ...newTask, id: e.target.value })}
        /> */}
        <select
          className="border px-2 py-1"
          value={newTask.status}
          onChange={(e) => setNewTask({ ...newTask, status: e.target.value })}>
          <option>Pending</option>
          <option>Running</option>
          <option>Completed</option>
          <option>Failed</option>
        </select>
        <button
          className="bg-blue-500 text-white px-3 py-1 rounded"
          onClick={addTask} >
          Add Task
        </button>
      </div>
      <DataTable data={data} />
    </main>
    </SidebarInset>
  </SidebarProvider>
  );
}
