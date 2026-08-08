```text
         ┌───────────┐
         │   CLI     │
         └─────┬─────┘
               │
         ┌─────▼─────┐
         │  Usecase  │
         └─────┬─────┘
               │
     ┌─────────▼─────────┐
     │   SonarReport     │
     └─────────┬─────────┘
               │
     ┌─────────┴─────────┐
     ▼                   ▼
  REST API          GitLab Renderer
     │                   │
     ▼                   ▼
  Client             MR Comment
```