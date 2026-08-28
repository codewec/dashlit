import { readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const docsDirectory = fileURLToPath(new URL('../', import.meta.url));
const source = fileURLToPath(new URL('../../CHANGELOG.md', import.meta.url));
const destination = fileURLToPath(new URL('../changelog.md', import.meta.url));
const russianDestination = fileURLToPath(new URL('../ru/changelog.md', import.meta.url));

const changelog = await readFile(source, 'utf8');
const russianChangelog = changelog.replace(/^# Changelog\s*/, '# История изменений\n\n');

await writeFile(destination, changelog);
await writeFile(russianDestination, russianChangelog);
console.log(`Synced CHANGELOG.md into ${destination.slice(docsDirectory.length)}`);
