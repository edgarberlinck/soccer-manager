# Plano de Testes das Simulacoes

Este documento define o que deve ser testado nas simulacoes do projeto, com foco em confiabilidade do motor de jogo, estabilidade do modo debug, consistencia de calendario e processamento concorrente no scheduler.

## Objetivo

Garantir que:

- A simulacao de partida gera resultados plausiveis e nao viciados.
- O modo debug/manual permanece reproduzivel quando seed fixa e observavel para analise.
- O artefato gerado por `make simulation` contenha todos os dados esperados para backend e frontend.
- O calendario por tick e o scheduler concorrente mantenham consistencia sob carga e ociosidade.

## Escopo

Coberto neste plano:

- Simulacao por tick e resultado final da partida.
- Simulacao debug com snapshots de campo e eventos.
- Estatisticas por jogador e atribuicao de autor do lance.
- Substituicoes automaticas por desgaste e rendimento.
- Calendario esportivo/administrativo e agenda por tick.
- Scheduler com fila e workers.
- Integridade do JSON em `tmp/simulation-output.json`.

Nao coberto neste plano:

- Performance de infraestrutura em ambiente de producao.
- Testes de UI frontend.
- Persistencia em banco e endpoints HTTP da API.

## Comandos Base

### Suite completa

```bash
go test ./...
```

### Pacotes criticos

```bash
go test ./simulation ./cmd/simulate-debug ./internal/domain/calendar ./internal/infrastructure/scheduler
```

### Geracao de artefato manual

```bash
make simulation
```

### Simulacao com seed fixa

```bash
go run ./cmd/simulate-debug -seed 21 ./simulation/testdata/manual_debug/home_debug.json ./simulation/testdata/manual_debug/away_debug.json
```

## Checklist Obrigatorio por Area

## 1. Motor de partida (simulation/play_match)

Validar:

- Executa 90 ticks sem panic.
- Score final e consistente com eventos de gol.
- Distribuicao de resultados nao fica concentrada em um unico placar (exemplo: 7x0 recorrente).
- Troca de posse e zonas de bola acontece ao longo da partida.
- Fator de skill do time nao introduz vies numerico (underflow/overflow).

Cenarios minimos:

- Seed fixa para reproducao de regressao.
- Multiplas seeds para variacao esperada.
- Times equilibrados e times desbalanceados.

Criterio de aceite:

- Testes passam e nao ha padrao anormal de placar repetitivo em amostra curta de seeds.

## 2. Modo debug/manual (simulation/debug_match)

Validar:

- Gera 90 snapshots (`regulationTicks`).
- Render inclui cabecalho, legenda e campo com marcadores.
- Matriz de campo tem dimensoes esperadas e metadata de ocupantes.
- Estados taticos (posse, zona, modo casa/fora) presentes por tick.
- Historico de substituicoes e stamina minima por time evoluem de forma coerente.

Cenarios minimos:

- Times com exatamente 11 jogadores.
- Times com banco (12+ jogadores) para forcar substituicao automatica.
- Seed fixa para reproducao de casos reportados.

Criterio de aceite:

- Nenhum tick ausente, sem campos nulos criticos e com comportamento estavel na reproducao por seed.

## 3. Estatisticas por jogador e autoria do evento

Validar:

- Descricao de evento enriquecida com nome do jogador quando aplicavel (`Jogador: <nome>`).
- Campo de ator do evento presente no snapshot (`EventActorID`, `EventActorName`, `EventActorTeam`) quando evento tiver autor.
- Estatisticas acumulam ao longo da partida por jogador:
  - `movement`
  - `touches`
  - `correct_touches`
  - `long_passes`
  - `shots_on_goal`
  - `fouls`
- Ordenacao e snapshot de elenco continuam estaveis apos inclusao de stats.

Cenarios minimos:

- Evento de passe curto/longo.
- Evento defensivo (desarme/interceptacao/falta).
- Evento ofensivo com finalizacao/gol.

Criterio de aceite:

- Pelo menos um jogador com stats > 0 durante a partida e logs com autoria quando esperado.

## 4. Substituicoes automaticas

Validar:

- Substituicao ocorre quando ha banco elegivel e condicoes de fadiga/rendimento.
- Motivo da substituicao faz sentido (`desgaste`, `baixo rendimento`, etc).
- Jogador substituido fica inativo e substituto entra ativo.
- Nao excede limite configurado de substituicoes.

Cenarios minimos:

- Jogador titular com stamina inicial muito baixa.
- Reserva com encaixe posicional razoavel.
- Caso sem banco para confirmar que nao quebra.

Criterio de aceite:

- Sem inconsistencias de estado (dois jogadores na mesma vaga por erro de estado, jogador inativo recebendo evento, etc).

## 5. Relatorio final e artefato JSON (cmd/simulate-debug)

Validar:

- `tmp/simulation-output.json` e gerado sempre que `make simulation` roda com sucesso.
- JSON contem:
  - metadados da simulacao (`seed`, `generated_at`, `match_id`, placar)
  - `calendar` completo
  - `snapshots`
  - `performance_summary` com top 3 por metrica para casa e fora
- `performance_summary` usa dados do ultimo snapshot e ranking em ordem decrescente por valor.

Cenarios minimos:

- Escrita em pasta inexistente (deve criar diretorios).
- Caminho de saida customizado com `-out`.
- `-calendar-tick-seconds` valido e impacto no calendario.

Criterio de aceite:

- JSON parseavel, campos obrigatorios presentes e sem divergencia entre score final e ultimo snapshot.

## 6. Calendario por tick (internal/domain/calendar)

Validar:

- Conversao de tempo real para tick do dia (`TickAt`) coerente.
- `DayBounds`, `TickRangeForDay` e agenda atual sem intervalos invalidos.
- Entradas esportivas e administrativas coexistem corretamente.
- Planejamento de simulacao (`PlanMatchSimulation`) respeita limites de lote/paralelismo.

Cenarios minimos:

- Inicio, meio e fim de dia.
- Janela de transferencia cobrindo todo o dia.
- Partida perto do limite do dia para testar truncagem de fim.

Criterio de aceite:

- Agenda consistente, sem ticks negativos e sem inversao de intervalos.

## 7. Scheduler concorrente (internal/infrastructure/scheduler)

Validar:

- Loop periodico processa ticks sem deadlock.
- Modelo producer/consumer respeita tamanho de fila e numero de workers.
- Janela ociosa dispara hook esperado.
- Sem perda de trabalho em cargas pequenas e medias.

Cenarios minimos:

- `SIMULATION_WORKER_POOL_SIZE` baixo (1) e maior (2+).
- `SIMULATION_QUEUE_SIZE` pequeno para testar pressao de fila.
- Tick sem partidas (idle) e com partidas ativas.

Criterio de aceite:

- Testes do scheduler estaveis, sem corrida aparente e sem travamento.

## 8. Regressao e sanidade operacional

Executar antes de merge de qualquer alteracao de simulacao:

```bash
go test ./...
make simulation
```

Conferencias manuais minimas:

- Console mostra eventos com jogador quando aplicavel.
- Console mostra bloco `RESUMO DE PERFORMANCE (TOP 3)`.
- `tmp/simulation-output.json` atualizado e com `performance_summary`.

## Matriz Rapida de Aceite

| Area             | Gate minimo                                          |
| ---------------- | ---------------------------------------------------- |
| Motor de partida | `go test ./simulation` verde                         |
| Debug/manual     | snapshots completos + render estavel                 |
| Stats/autoria    | logs com `Jogador:` + stats acumulados               |
| Substituicoes    | ao menos um caso coberto em teste                    |
| Relatorio JSON   | `make simulation` gera artefato valido               |
| Calendario       | testes de `internal/domain/calendar` verdes          |
| Scheduler        | testes de `internal/infrastructure/scheduler` verdes |

## Frequencia Recomendada

- Durante desenvolvimento local: rodar pacote alvo + `make simulation`.
- Antes de abrir PR: rodar checklist de regressao completo.
- Antes de merge na branch principal: executar novamente `go test ./...` com workspace limpo.

## Sinais de Alerta

Investigar imediatamente se ocorrer:

- Placar repetido de forma anormal em seeds diferentes.
- Logs sem autor de evento em todos os ticks.
- Stats zeradas para todos os jogadores no fim da partida.
- Substituicoes inconsistentes (estado ativo/inativo errado).
- Divergencia entre score final do relatorio e ultimo snapshot.
- Calendario com tick fora de faixa do dia.
- Scheduler travando ou acumulando backlog sem drenagem.

## Evolucao do Plano

Quando nova feature de simulacao for adicionada, atualizar este documento com:

- O que muda no comportamento esperado.
- Novos cenarios obrigatorios.
- Novos campos obrigatorios de artefato.
- Novos gates de regressao.
