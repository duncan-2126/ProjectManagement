# LOCAL_SETUP.md - Countdowns (feature/CCDC)

## Prerequisites

- **Node.js** 18.x or later
- **npm** or **pnpm**
- **Git**

## Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/duncan-2126/Countdowns.git
cd Countdowns
git checkout feature/CCDC
```

### 2. Install Dependencies

```bash
npm install
# or
pnpm install
```

### 3. Start Development Server

```bash
npm run dev
```

The application will start at **http://localhost:3000**

### 4. Build for Production (Optional)

Single configuration:
```bash
NEXT_PUBLIC_DISPLAY_TYPE=centralLeft NEXT_PUBLIC_SCHEDULE_TYPE=snowShow npm run build
```

All configurations:
```bash
npm run build:all
```

Output is in the `out/` directory.

## Known Issues

### ⚠️ Schedule Dates Are Expired
The schedule data in `src/data/scheduleData.ts` contains dates from November-December **2025**. As of February 2026, all events are in the past, so the countdown will show "No events scheduled."

**Fix**: Update the schedule dates to future dates (e.g., 2026 or later).

### ⚠️ Missing Video Assets
Video files expected at `/assets/{scheduleType}/{displayType}.mp4` do not exist. Video backgrounds will 404.

**Fix**: Add video files to `public/assets/{snowShow,lightShow}/` directories.

### ⚠️ Missing Custom Font
The font file `public/font/ProximaNova-Bold.otf` is referenced but not present.

**Fix**: Add the font file or remove the font preload reference in `src/app/layout.tsx`.

### ⚠️ URL Parameters Don't Work
The README claims `?display=centralLeft` and `?schedule=lightShow` work for development, but only environment variables are used.

## Display Types & Resolutions

| Display Type | Resolution |
|--------------|------------|
| centralLeft | 1440x675 |
| centralRight | 1440x675 |
| sElevator | 1920x480 |
| ddTec | 1080x1920 |
| sportCheck | 1632x816 |
| serviceDesk | 1496x340 |
| bayDundas | 603x350 |
| atrium | 3520x1980 |

## Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build with env vars |
| `npm run build:all` | Build all 15 configurations |
| `npm run test` | Run Jest tests |
| `npm run clean` | Remove out/ and .next/ |

## Environment Variables

- `NEXT_PUBLIC_DISPLAY_TYPE` - Display layout (e.g., centralLeft)
- `NEXT_PUBLIC_SCHEDULE_TYPE` - Schedule type (snowShow or lightShow)
