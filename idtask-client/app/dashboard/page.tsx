"use client"

import { AppSidebar } from "@/components/app-sidebar"
import { ChartAreaInteractive } from "@/components/chart-area-interactive"
import { SectionCards } from "@/components/section-cards"
import { SiteHeader } from "@/components/site-header"
import {
  SidebarInset,
  SidebarProvider,
} from "@/components/ui/sidebar"

import data from "./data.json"
import { useEffect, useState } from "react"
import { apiClient } from "@/lib/api-client"

export default function Page() {

  const [schedulers, setSchedulers] = useState([]);
  const [workers, setWorkers] = useState([]);

  useEffect(() => {
    apiClient.post("/scheduler/status", {})
      .then(setSchedulers)
      .catch((err) => console.error("Failed to load schedulers:", err));

    apiClient.post("/worker/status", {})
      .then(setWorkers)
      .catch((err) => console.error("Failed to load workers:", err));
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
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              <div className="px-4 lg:px-6">
                <h2 className="text-lg font-semibold mb-2">Schedulers</h2>
              </div>
              <SectionCards data={schedulers} />
              <div className="px-4 lg:px-6">
                <h2 className="text-lg font-semibold mb-2">Workers</h2>
              </div>
              <SectionCards data={workers} />
              <div className="px-4 lg:px-6">
                <ChartAreaInteractive />
              </div>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
