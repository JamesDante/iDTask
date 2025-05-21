import { IconTrendingDown, IconTrendingUp } from "@tabler/icons-react"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardAction,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { format } from "date-fns/format";

interface SchedulerCard {
  id: string;
  status: string;
  isLeader: string;
  heart_beat: string;
}

interface SchedulerCardPros {
  data: SchedulerCard[];
}

export function SectionWorkerCards({ data }: SchedulerCardPros) {
  return (
    <div className="*:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      {data.map((s: any, i) => (
            <Card className="@container/card" key={s.id}>
              <CardHeader>
                <CardDescription>worker status</CardDescription>
                <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
                  {s.id}
                </CardTitle>
                <CardAction>
                  <Badge variant="outline">
                    <IconTrendingUp />
                    {s.status}
                  </Badge>
                </CardAction>
              </CardHeader>
              <CardFooter className="flex-col items-start gap-1.5 text-sm">
                <div className="text-muted-foreground">
                  Heartbeat At: {format(new Date(s.heart_beat), "H:mm:ss")}
                </div>
              </CardFooter>
            </Card>
      ))}


    </div>
  )
}
