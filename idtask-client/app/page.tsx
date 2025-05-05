// app/page.tsx
import { TaskTable } from "@/components/task-table";

const sampleTasks = [
  { id: "task-1", status: "Pending", duration: 12 },
  { id: "task-2", status: "Running", duration: 30 },
  { id: "task-3", status: "Failed", duration: 7 },
];

export default function HomePage() {
  return (
    <main className="p-6">
      <h1 className="text-2xl font-bold mb-4">Task List</h1>
      <TaskTable data={sampleTasks} />
    </main>
  );
}
