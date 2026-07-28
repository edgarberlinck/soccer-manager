import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Calendar, Home, Plane, Trophy, Clock } from 'lucide-react'

interface Match {
  id: string
  home_club_id: string
  away_club_id: string
  status: string
  home_score?: number
  away_score?: number
  is_home: boolean
  opponent_id: string
  created_at: string
  finished_at?: string
}

interface CalendarStats {
  total_matches: number
  home_matches: number
  away_matches: number
  completed_matches: number
  pending_matches: number
}

interface CalendarData {
  club_id: string
  matches: Match[]
  stats: CalendarStats
}

async function fetchCalendar(clubId?: string): Promise<CalendarData> {
  const url = clubId
    ? `/api/calendar/clubs/${clubId}`
    : '/api/calendar/season'
  
  const response = await fetch(url, {
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`,
    },
  })
  
  if (!response.ok) {
    throw new Error('Failed to fetch calendar')
  }
  
  return response.json()
}

export function CalendarView({ clubId }: { clubId?: string }) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['calendar', clubId],
    queryFn: () => fetchCalendar(clubId),
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  if (error) {
    return (
      <Card className="border-red-200">
        <CardHeader>
          <CardTitle className="text-red-600">Erro ao carregar calendário</CardTitle>
          <CardDescription>{error.message}</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Estatísticas */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
        <StatsCard
          title="Total"
          value={data?.stats.total_matches || 0}
          icon={<Calendar className="h-4 w-4" />}
          color="blue"
        />
        <StatsCard
          title="Em Casa"
          value={data?.stats.home_matches || 0}
          icon={<Home className="h-4 w-4" />}
          color="green"
        />
        <StatsCard
          title="Fora"
          value={data?.stats.away_matches || 0}
          icon={<Plane className="h-4 w-4" />}
          color="purple"
        />
        <StatsCard
          title="Concluídas"
          value={data?.stats.completed_matches || 0}
          icon={<Trophy className="h-4 w-4" />}
          color="yellow"
        />
        <StatsCard
          title="Pendentes"
          value={data?.stats.pending_matches || 0}
          icon={<Clock className="h-4 w-4" />}
          color="gray"
        />
      </div>

      {/* Lista de Partidas */}
      <Card>
        <CardHeader>
          <CardTitle>Próximas Partidas</CardTitle>
          <CardDescription>
            {data?.matches.length || 0} partidas agendadas
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {data?.matches.slice(0, 10).map((match) => (
              <MatchCard key={match.id} match={match} />
            ))}
            {!data?.matches.length && (
              <p className="text-center text-gray-500 py-8">
                Nenhuma partida agendada
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

interface StatsCardProps {
  title: string
  value: number
  icon: React.ReactNode
  color: 'blue' | 'green' | 'purple' | 'yellow' | 'gray'
}

function StatsCard({ title, value, icon, color }: StatsCardProps) {
  const colors = {
    blue: 'bg-blue-50 text-blue-600 border-blue-200',
    green: 'bg-green-50 text-green-600 border-green-200',
    purple: 'bg-purple-50 text-purple-600 border-purple-200',
    yellow: 'bg-yellow-50 text-yellow-600 border-yellow-200',
    gray: 'bg-gray-50 text-gray-600 border-gray-200',
  }

  return (
    <Card className={colors[color]}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardDescription className="text-xs font-medium">
            {title}
          </CardDescription>
          {icon}
        </div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
      </CardContent>
    </Card>
  )
}

function MatchCard({ match }: { match: Match }) {
  const getStatusBadge = (status: string) => {
    const variants = {
      pending: { variant: 'secondary' as const, label: 'Agendada' },
      in_progress: { variant: 'default' as const, label: 'Ao Vivo' },
      finished: { variant: 'outline' as const, label: 'Finalizada' },
    }
    const config = variants[status as keyof typeof variants] || variants.pending
    return <Badge variant={config.variant}>{config.label}</Badge>
  }

  const date = new Date(match.created_at)
  const formattedDate = date.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })

  return (
    <div className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors">
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-2">
          {match.is_home ? (
            <Home className="h-4 w-4 text-green-600" />
          ) : (
            <Plane className="h-4 w-4 text-purple-600" />
          )}
          <span className="text-sm text-gray-500">{formattedDate}</span>
        </div>
        <div className="font-medium">
          {match.is_home ? 'vs Oponente' : '@ Oponente'}
        </div>
      </div>

      <div className="flex items-center space-x-4">
        {match.home_score !== undefined && match.away_score !== undefined && (
          <div className="text-lg font-bold">
            {match.home_score} - {match.away_score}
          </div>
        )}
        {getStatusBadge(match.status)}
      </div>
    </div>
  )
}
