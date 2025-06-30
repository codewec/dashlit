import { dataPath, filePath } from '$lib/server/helper';
import { type RequestHandler } from '@sveltejs/kit';
import fs from 'node:fs/promises';

export const GET: RequestHandler = async ({ url }) => {
	const fileName = `${dataPath()}/custom.css`;
	const data = await fs.readFile(fileName, { encoding: 'utf8' }).catch(() => {
		console.log(`File ${fileName} not found`);
		return '';
	});

	return new Response(String(data), {
		headers: {
			'Content-type': 'text/css'
		}
	});
};
