// Entry for the V8 startup snapshot and the SEA: module init (including `yaml`) happens at build time,
// `run` at startup. Bundled to CommonJS by run.ts.
import { startupSnapshot } from 'node:v8';
import { run } from './hook-lib.ts';

if (startupSnapshot.isBuildingSnapshot()) startupSnapshot.setDeserializeMainFunction(run);
else run();
