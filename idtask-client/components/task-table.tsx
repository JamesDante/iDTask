import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"

interface Task {
  id: string;
  status: string;
  duration: number;
}

interface TaskTableProps {
  data: Task[];
}

export  function TaskTable({ data }: TaskTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableCell>任务 ID</TableCell>
          <TableCell>状态</TableCell>
          <TableCell>执行时间</TableCell>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.map(task => (
          <TableRow key={task.id}>
            <TableCell>{task.id}</TableCell>
            <TableCell>{task.status}</TableCell>
            <TableCell>{task.duration}s</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}