# apps/mobile

Placeholder for the Expo/React Native mobile app — built in **BLUEPRINT.md Phase 7 (Offline-First & Mobile)**, after the web app and API are stable, since the mobile app shares its data layer and offline-sync engine with `apps/web`.

To scaffold when that phase starts:

```bash
cd apps
npx create-expo-app@latest mobile --template
```

Then wire in `@zarishlog/business-logic` and `@zarishlog/data-models` exactly as `apps/web` does, plus the offline sync engine (RxDB/PowerSync + SQLite) described in `ARCHITECTURE.md` §6.
