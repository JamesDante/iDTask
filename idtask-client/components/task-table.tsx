import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"

interface Task {
  id: string;
  status: string;
  type: string;
  assgin: string;
  duration: number;
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
          <TableCell>Task ID</TableCell><
            TableCell>Type</TableCell>
          <TableCell>Status</TableCell>
          <TableCell>Duration</TableCell>
          <TableCell>Assign</TableCell>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.map(task => (
          <TableRow key={task.id}>
            <TableCell>{task.id}</TableCell>
            <TableCell>{task.type}</TableCell>
            <TableCell>{task.status}</TableCell>
            <TableCell>{task.duration}s</TableCell>
            <TableCell>{task.executed_by.Valid ? task.executed_by.String : "N/A"}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}