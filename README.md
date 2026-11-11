# git-koi-aquarium 🐟

[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-git--koi--aquarium-blue?logo=github&style=flat-square)](https://github.com/marketplace/actions/git-koi-aquarium)
[![Go Reference](https://pkg.go.dev/badge/github.com/reginaldalfret/git-koi-aquarium.svg)](https://pkg.go.dev/github.com/reginaldalfret/git-koi-aquarium)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

Draw an animated, pixel-art koi fish and live moving SVG aquarium from any GitHub account's public activity as a looping GIF, still PNG, live aquarium, and preview page for a profile README.

This repository is a native **Go rewrite** of [`git-koi`](https://github.com/yalongwastaken/git-koi), providing high-speed execution, deterministic pixel art generation, and zero external dependencies using Go's standard library (`image/gif`, `image/png`, `net/http`).

---

## How It Works

Every aspect of the koi's appearance and motion translates public GitHub activity into biological traits:

| GitHub Metric | Koi Biological Trait | Visual Manifestation |
|---|---|---|
| **Language Shares** | **Pigments & Coat** | Primary languages map to nearest hues: Beni (Red), Sumi (Black), Ki (Yellow), Orenji (Orange), Ai (Blue). |
| **Total Contributions** | **Growth Curve** | Scaled across logarithmic thresholds (500, 1,500, 12,750, 100,000) governing body length, head shape, and fin proportions. |
| **Weekly Streaks** | **Gin-Rin & Fin Marks** | Active streaks color fin tips; record streaks unlock *Gin-rin* sparkling reflective scales and vivid pigments. |
| **Recent Activity** | **Swimming Behavior** | Selects behavior: **Deep Rest** (idle > 60 days), **Rest** (idle > 14 days), **Surface Feeding** (first contribution of week with ripple rings), **Dart** (crossed milestones or busy week), or **Cruise** (steady lap). |

---

## Key Advantages of the Go Port

- **Zero Third-Party Image Libraries:** Uses Go's native `image/gif`, `image/png`, and `image/draw` standard libraries.
- **Blazing Fast Compilation & Execution:** Full 120-frame animation rendering completes in under **0.55 seconds**.
- **Compact Standalone Binaries:** Single static binary with zero external runtime requirements.
- **Deterministic Output:** Bit-for-bit reproducible pixels and frame timing matching original mathematical models.

---

## CLI Usage

### 1. Collect Account Snapshot

Collect public owned and contributed repositories, language colors, and continuous daily contribution history:

```bash
git-koi collect octocat --output snapshot.json
```

*Requires GitHub CLI (`gh`) authenticated with `gh auth login` or `GH_TOKEN`.*

### 2. Render Koi Animations

Render `koi.gif` (swimming animation loop), `koi.png` (still frame), and `preview.html`:

```bash
git-koi render snapshot.json --output-dir ./dist --body asagi
```

Supported body grounds: `white` *(default)*, `black`, `asagi`, `tea`, `gold`, `platinum`.

### 3. Render Live Moving Aquarium (6 Fishes)

Render a self-contained, animated SVG aquarium (`aquarium.svg`) with all 6 native koi varieties swimming along organic, non-repeating Bézier curves with dynamic banking, light caustics, and water ripples:

```bash
git-koi aquarium snapshot.json --output-dir ./dist
```

---

## GitHub Actions Integration

Add nightly generation to your profile README repository:

```yaml
name: Generate Profile Koi

on:
  schedule:
    - cron: '15 0 * * *' # Nightly after UTC midnight
  workflow_dispatch:

jobs:
  draw-koi:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4

      - name: Draw GitHub Profile Koi Aquarium
        uses: reginaldalfret/git-koi-aquarium@v1
        with:
          login: ${{ github.repository_owner }}
          body: all
          output-dir: koi
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Commit & Push Changes
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add koi/
          git diff --quiet && git diff --staged --quiet || git commit -m "chore: update daily profile koi [skip ci]"
          git push
```

---

## Architectural Modules

- [`pkg/history`](./pkg/history): Account snapshots, continuous ISO week normalization, and repository language share calculations.
- [`pkg/koi`](./pkg/koi): Development derivation, streak tracking, and logarithmic growth curves.
- [`pkg/behavior`](./pkg/behavior): Daily behavioral state machine heuristics.
- [`pkg/coat`](./pkg/coat): Pigment clustering, organic patch sampling, and Kohaku/Asagi variety layouts.
- [`pkg/markings`](./pkg/markings): Streak fin strength, color vibrancy shifts, and Gin-rin glimmer scale generation.
- [`pkg/motion`](./pkg/motion): 15 FPS kinematics, spine wave mechanics, and closed-loop swim curves.
- [`pkg/render`](./pkg/render): Native 256x256 pixel art rasterizer, fin rays, mouth mechanics, and water ripples.
- [`pkg/output`](./pkg/output): 2x scale (512x512) nearest-neighbor scaling, bounding-box palette quantization, GIF/PNG encoding.
- [`pkg/collect`](./pkg/collect): GitHub GraphQL client querying `contributionsCollection` and user repositories.

---

## License

MIT License.
