# Dimension Reduction Strategy

## Goal
Shrink 768-d embedding vectors to lower-dimensional representations without materially degrading retrieval or model performance.

## Techniques
1. **PCA / SVD**
   - Fit on representative embedding batches.
   - Retain 256–384 components; store projection matrix and explained variance.
2. **Random Projection**
   - Use sparse Johnson–Lindenstrauss transforms for fast, low-overhead reductions.
   - Benchmark against PCA for recall retention.
3. **Autoencoders**
   - Train shallow bottleneck networks (e.g., 768→256→768) on historical embeddings.
   - Monitor reconstruction error and retrieval quality.
4. **Product Quantization (OPQ + IVF-PQ)**
   - Integrate with FAISS for large-scale vector search to reduce storage + accelerate queries.

## Workflow
1. **Data Prep**
   - Gather balanced sample across sessions and chunk types.
   - Standardize embeddings (mean-center, optionally whiten).
2. **Model Training / Fitting**
   - Run PCA baseline and evaluate explained variance.
   - Experiment with autoencoder and OPQ configurations when higher compression needed.
3. **Evaluation**
   - Compute recall@k on labeled similarity pairs before/after reduction.
   - Track absolute cosine similarity differences; set acceptable thresholds (e.g., Δcos ≤ 0.05 mean).
4. **Deployment**
   - Store projection parameters in versioned artifact registry.
   - Apply reduction during embedding pipeline post-processing with opt-in flag.
   - For vector DBs, load reduced vectors and keep mapping to original IDs.

## Deliverables
- Reduction playbook with benchmarking results.
- Projection matrices / autoencoder weights stored under `models/dimension_reduction/`.
- Integration tests ensuring embedding pipeline supports reduced and full-dimension modes.

## Risks & Mitigations
- **Quality Loss**: mitigate by monitoring recall metrics and retaining ability to re-run with full vectors.
- **Drift**: schedule periodic retraining when embedding distributions shift.
- **Operational Complexity**: encapsulate transformation logic in shared library to avoid duplicated math.
