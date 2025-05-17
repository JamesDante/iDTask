import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import { format, parseISO } from "date-fns"

interface Task {
  id: string;
  status: string;
  type: string;
  assgin: string;
  duration: number;
  created_at: string;
  executed_at: string;
  executed_by: { String: string; Valid: boolean };
}

interface TaskTableProps {
  data: Task[];
}

export  function TaskTable({ data }: TaskTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableCell>Task ID</TableCell>
          <TableCell>Type</TableCell>
          <TableCell>Status</TableCell>
          <TableCell>Created At</TableCell>
          <TableCell>Assign</TableCell>
          <TableCell>Executed At</TableCell>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.map(task => (
          <TableRow key={task.id}>
            <TableCell>{task.id}</TableCell>
            <TableCell>{task.type}</TableCell>
            <TableCell>{task.status}</TableCell>
            <TableCell>{format(parseISO(task.created_at), "yyyy-MM-dd HH:mm:ss")}</TableCell>
            <TableCell>{task.executed_by.Valid ? task.executed_by.String: 'N/A'}</TableCell>
            <TableCell>{task.executed_at?format(parseISO(task.executed_at), "yyyy-MM-dd HH:mm:ss"):'N/A'}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}