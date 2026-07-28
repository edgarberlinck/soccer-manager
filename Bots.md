# Sistema de Bots

## Visão Geral

O sistema de bots adiciona 2000 **jogadores controlados por IA** ao jogo, indistinguíveis de jogadores humanos. Cada bot tem seu próprio time, perfil e toma decisões automaticamente. 

**Importante:** Bots são **oponentes** que permitem que jogadores humanos sempre tenham com quem jogar, mesmo quando não há outros jogadores online. Eles são tratados exatamente como jogadores reais no sistema - não há diferenciação visível para o usuário final.

## Estratégias

Cada bot possui uma das três estratégias:

### 1. **Conservador** (~667 bots)
- Prioriza treinos leves (Soft)
- Descansa jogadores quando status físico < 60
- Foco em preservar a condição física dos jogadores
- Menos arriscado, desenvolvimento mais lento

### 2. **Equilibrado** (~667 bots)
- Mix balanceado entre treinos leves, médios e intensos
- Descansa jogadores quando status físico < 40
- Abordagem moderada ao desenvolvimento
- Equilíbrio entre risco e recompensa

### 3. **Agressivo** (~666 bots)
- Prioriza treinos intensos (Intense)
- Apenas descansa jogadores quando status físico < 25
- Desenvolvimento rápido com alto risco de fadiga
- Máximo ganho, máximo desgaste

## Funcionalidades

### Jogadores Indistinguíveis
- Bots aparecem como jogadores normais em todos os lugares
- Têm perfis, times, histórico de partidas
- Sem indicação visual de que são IA
- Fazem parte do mesmo ranking/liga que jogadores reais

### Matchmaking Automático
- Quando um jogador busca uma partida, o sistema pode conectá-lo com um bot
- Bots aceitam desafios automaticamente
- Escalam o melhor time disponível para cada partida
- Garantem que sempre há oponentes disponíveis

## Arquitetura

### Componentes

1. **Migration** (`00008_add_bot_support.sql`)
   - Adiciona campos `is_bot` e `bot_strategy` na tabela users
   - Indexa usuários bot para queries eficientes

2. **Queries SQL** (`bot.sql`)
   - `CreateBotUser`: Cria um usuário bot
   - `GetAllBots`: Lista todos os bots
   - `GetBotsByStrategy`: Filtra bots por estratégia
   - `GetBotClubs`: Lista clubes pertencentes a bots

3. **Domínio** (`internal/domain/bot/bot.go`)
   - Lógica de decisão dos bots
   - Estratégias de treino e descanso
   - Seleção de melhores jogadores

4. **Serviço** (`internal/services/bot_service.go`)
   - Orquestra ações dos bots
   - Executa treinos e atualizações
   - Gerencia scheduler periódico

5. **Comandos**
   - `create-bots`: Cria os 2000 bots no banco
   - `bot-scheduler`: Roda ações dos bots periodicamente

## Como Usar

### 1. Criar os Bots

```bash
cd api
go run cmd/create-bots/main.go
```

Isso criará:
- 2000 usuários bot
- 2000 clubes (um para cada bot)
- Distribuição balanceada entre as 3 estratégias

### 2. Rodar o Scheduler

```bash
go run cmd/bot-scheduler/main.go
```

O scheduler:
- Executa imediatamente ao iniciar
- Repete a cada 30 minutos
- Pode ser interrompido com Ctrl+C
- Logs detalhados das ações

### 3. Rodar a Migration

```bash
# Usando goose ou sua ferramenta de migration preferida
goose -dir internal/infrastructure/database/migrations postgres "sua-connection-string" up
```

### 4. Gerar o código sqlc

```bash
# Após adicionar as novas queries
sqlc generate
```

## Configurações

### Intervalo do Scheduler
Padrão: 30 minutos

Para alterar, edite `cmd/bot-scheduler/main.go`:
```go
botService.StartBotScheduler(ctx, 15*time.Minute) // 15 minutos
```

### Número de Jogadores por Rodada
Padrão: 1-5 jogadores aleatórios (máximo 5)

Para alterar, edite `internal/services/bot_service.go`:
```go
if numToTrain > 10 { // aumenta para 10
    numToTrain = 10
}
```

## Logs e Monitoramento

O sistema gera logs informativos:

```
Executando ações para 2000 bots...
Progresso: 100 bots criados...
Progresso: 200 bots criados...
...
✅ Processo concluído!
Bots criados com sucesso: 2000
Falhas: 0

Distribuição:
- Conservadores: ~667
- Equilibrados: ~667
- Agressivos: ~666
```

## Integrações Futuras

O sistema está preparado para:
- [ ] Partidas automáticas entre bots
- [ ] Transferências entre bots
- [ ] Ligas e competições exclusivas para bots
- [ ] Sistema de reputação/ranking
- [ ] Eventos aleatórios (lesões, descobertas, etc)

## Observações Técnicas

- Bots não fazem login (password hash vazio)
- Ações são executadas em background
- Não há impacto na performance do jogo principal
- Bots usam as mesmas regras de engine que jogadores reais
- Todos os bots têm `active = true` por padrão
