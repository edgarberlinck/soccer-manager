import { createFileRoute } from '@tanstack/react-router'
import { CalendarView } from '@/components/calendar/CalendarView'

export const Route = createFileRoute('/calendar/')({
  component: CalendarPage,
})

function CalendarPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Calendário da Temporada</h1>
      <CalendarView />
    </div>
  )
}
