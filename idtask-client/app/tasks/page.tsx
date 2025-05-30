// app/page.tsx
"use client"

import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { apiClient } from "@/lib/api-client";
import { useEffect, useState } from "react";

import { TaskTable } from "@/components/task-table";
import { taskTypes } from "@/lib/task-types";

import data from "../dashboard/data.json"

export default function HomePage() {
  const [tasks, setTasks] = useState(data);
  const [newTask, setNewTask] = useState({ type: "0", status: "Pending" });
  const [currentPage, setCurrentPage] = useState(1);
  const [total, setTotal] = useState(1);
  const tasksPerPage = 10;

  const [tasksList, setTasksList] = useState([]);

  const indexOfLastTask = currentPage * tasksPerPage;
  //const indexOfFirstTask = indexOfLastTask - tasksPerPage;
  const currentTasks = Array.isArray(tasksList) ? tasksList : [];
  const totalPages = Math.ceil((total || 0) / tasksPerPage);

  const loadTaskList= (page = currentPage)=>{
    apiClient.post("/tasks/list", {
        page: page,
        page_size: 10,
      })
      //.then(setTasksList)
      .then((res) => {
        if (res.status === "OK") {
          setTasksList(res.data.list_data);
          setTotal(res.data.total);
        } else {
          console.error("Failed to load tasks list:", res.error || "Unknown error");
        }
      })
      .catch((err) => console.error("Failed to load tasks list:", err));
  }

  const addTask = async () => {
    //if (!newTask.id) return;
    try {
      const response = await apiClient.post("/tasks", newTask)
      
      if (response.status === "OK") {
        loadTaskList();
      } else {
        console.error("Failed to add task:", response.error || "Unknown error");
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  useEffect(() => {
    loadTaskList();
  }, []);
    
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
          value={newTask.type}
          onChange={(e) => setNewTask({ ...newTask, type: e.target.value })}>
          {taskTypes.map((taskType) => (
            <option key={taskType.value} value={taskType.value}>
              {taskType.label}
            </option>
          ))}
       
        </select>
        <button
          className="bg-blue-500 text-white px-3 py-1 rounded"
          onClick={addTask} >
          Add Task
        </button>
      </div>
      <TaskTable data={currentTasks} />
<div className="flex justify-center mt-4 space-x-2">
  {Array.from({ length: totalPages }, (_, i) => i + 1)
    .filter(page => {
      const start = Math.max(1, currentPage - 4)
      const end = Math.min(totalPages, start + 9)
      return page >= start && page <= end
    })
    .map((page) => (
      <button
        key={page}
        onClick={() => {
          setCurrentPage(page)
          setTimeout(() => loadTaskList(page), 0)
        }}
        className={`px-3 py-1 rounded ${currentPage === page ? 'bg-blue-600 text-white' : 'bg-gray-200'}`}
      >
        {page}
      </button>
    ))}
</div>
    </main>
    </SidebarInset>
  </SidebarProvider>
  );
}
