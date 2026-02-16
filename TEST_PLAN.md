# Test Plan - Countdowns (feature/CCDC)

## Overview
This test plan covers the Countdowns Next.js application for digital signage countdown timers.

## Test Environment
- **Repository**: https://github.com/duncan-2126/Countdowns
- **Branch**: feature/CCDC
- **Tech Stack**: Next.js 14, React 18, TypeScript, Jest
- **Local URL**: http://localhost:3000

## Test Scenarios

### 1. Setup & Installation
- [ ] Clone repository and checkout feature/CCDC branch
- [ ] Run `npm install` successfully
- [ ] Verify no dependency errors

### 2. Development Server
- [ ] Run `npm run dev`
- [ ] Verify server starts on port 3000
- [ ] Verify main page loads without errors

### 3. Countdown Timer Functionality
- [ ] Verify countdown displays when schedule has future events
- [ ] Verify "No events scheduled" shows when all events are past
- [ ] Verify countdown updates every second
- [ ] Verify time remaining (hours, minutes, seconds) displays correctly

### 4. Schedule Types
- [ ] Test with `NEXT_PUBLIC_SCHEDULE_TYPE=snowShow`
- [ ] Test with `NEXT_PUBLIC_SCHEDULE_TYPE=lightShow`
- [ ] Verify correct schedule data loads for each type

### 5. Display Types
- [ ] Test each display type:
  - centralLeft, centralRight (1440x675)
  - sElevator (1920x480)
  - ddTec (1080x1920)
  - sportCheck (1632x816)
  - serviceDesk (1496x340)
  - bayDundas (603x350)
  - atrium (3520x1980)

### 6. Video Background
- [ ] Verify video background loads when assets available
- [ ] Verify graceful fallback when video missing (no crash)

### 7. Build Process
- [ ] Run `npm run build`
- [ ] Verify static export generates correctly
- [ ] Verify build completes without errors

### 8. Unit Tests
- [ ] Run `npm run test`
- [ ] Verify all tests pass

## Known Issues (Pre-Test)
1. Schedule dates are 2025 (expired) - FIXED to 2026
2. Video assets missing (expected - not critical for local dev)
3. Custom font missing (falls back gracefully)

## Acceptance Criteria
1. App starts without errors
2. Countdown timer displays and updates
3. Schedule data loads correctly
4. Build completes successfully
5. Tests pass

## Browser Automation (Optional)
For Playwright/Puppeteer validation:
```bash
npx playwright test
# or
npx puppeteer
```
