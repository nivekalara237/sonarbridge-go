```code

             SonarQube
                 │
  ┌──────────────┼────────────────┐
  ▼              ▼                ▼
Task        Quality Gate      Measures
  │                               │
  │                            Issues
  └──────────────┬────────────────┘
                 ▼
        ┌─────────────────┐
        │ AnalyzeService  │
        │                 │
        │  AGGREGATION    │
        └────────┬────────┘
                 ▼
           sonar.Report
                 │
        ┌────────┴─────────┐
        ▼                  ▼
  SaveReport()          Render()
        │                  │
        ▼                  ▼
    Repository          Markdown
                           │
                           ▼
                      GitLab MR


```


`caycStatus` signifie **Clean As You Code Status** — c'est le statut de conformité d'un projet par rapport à la politique **Clean as You Code** de Sonar.

---

## Concept

Clean as You Code (CaYC) est la philosophie Sonar qui dit : **ne t'occupe pas de la dette existante, concentre-toi sur le nouveau code**. `caycStatus` indique si ton projet est correctement configuré pour appliquer cette politique sur le new code.

---

## Valeurs possibles

| Valeur | Signification |
|---|---|
| `compliant` | Le quality gate est bien configuré CaYC — toutes les conditions portent sur le **new code** |
| `non-compliant` | Le quality gate n'est pas conforme CaYC — des conditions portent sur le **overall code** |
| `over-compliant` | Le quality gate est CaYC mais avec des conditions **trop strictes** (ex: coverage new code = 100%) |

---

## Où tu le trouves

Il apparaît dans la réponse de `/api/qualitygates/project_status` :

```json
{
  "projectStatus": {
    "status": "OK",
    "caycStatus": "compliant",
    "conditions": [...]
  }
}
```

Donc à ajouter dans ta struct :

```go
type QualityGateProjectStatus struct {
    Status            QualityGateStatus      `json:"status"`
    CaycStatus        string                 `json:"caycStatus"`
    Conditions        []QualityGateCondition `json:"conditions"`
    IgnoredConditions bool                   `json:"ignoredConditions"`
}
```

Ou avec un type dédié si tu veux être strict :

```go
type CaycStatus string

const (
    CaycCompliant    CaycStatus = "compliant"
    CaycNonCompliant CaycStatus = "non-compliant"
    CaycOverCompliant CaycStatus = "over-compliant"
)
```
