# Downstream Usage Strategy

## Goal
Leverage the structured action log for automation, analytics, and knowledge reuse across teams.

## Use Cases & Plans
1. **Fine-Tuning / Training Data**
   - Generate `(prompt -> action summary)` pairs from normalized data.
   - Split into train/validation by session, preserving temporal order.
   - Annotate with outcome labels (success/failure) to enable supervised instruction tuning.
2. **Pipeline Automation**
   - Create rules engine (e.g., Airflow DAG) watching PostgreSQL/JSONL for new `Bash(` actions with error details.
   - Trigger remediation runs or open tickets automatically.
3. **Observability Dashboards**
   - Load aggregates into BI tool (Metabase/Looker).
   - Visuals: actions per verb, error rates, top touched files, mean time between failures.
4. **Knowledge Base & Search**
   - Index embedding artifacts in vector store (Weaviate, Qdrant, pgvector).
   - Build query interface: ask “How was syntax highlighting fixed?” and retrieve full action bundles.

## Implementation Steps
1. **Data Contracts**
   - Publish schemas for consumers (JSON schema + SQL DDL).
   - Define SLAs for data availability (e.g., new sessions processed within 10 minutes).
2. **Tooling**
   - Provide SDK/CLI to fetch normalized sessions and embeddings.
   - Document sample notebooks for analysts.
3. **Governance & Access**
   - Apply role-based controls to production tables/vector indices.
   - Log query usage and track derived datasets.

## Success Measures
- Automated alerts generated for >90% of failed commands.
- ML dataset refresh cadence (weekly) met consistently.
- Search relevance evaluated via manual spot checks and recall metrics.
