import { Router } from "express";
import { lang, RunnerBuilder } from "./runners/runner";

export interface routerOpts {
	jsRunDir?: string
	jsTimeout?: number
}

const defJsTimeout = 30_000

export default function createRouter(opts: routerOpts): Router {
	const builder = new RunnerBuilder()
	if (opts.jsRunDir) {
		builder.addJs(opts.jsRunDir)
	}

	const runner = builder.bulid()

	const router = Router()

	router.get('/', (_, res) => {
		res.send('<h1> Heeeey </h1>')
	})

	router.get('/js', async (req, res) => {
		let { code } = req.query
		if (!code) {
			res.status(400).json(msgResp('code query param not provided'))
			return
		}

		code = code!.toString()

		try {
			const runres = await runner.run(
				lang.JS,
				code,
				opts.jsTimeout || defJsTimeout
			)

			res.status(200).json(runres)
		}
		catch (err) {
			console.error({ msg: 'exception during code run', err: err })
			res.status(500).json(msgResp('something went wrong'))
		}
	})

	return router
}

function msgResp(msg: string, details?: object): object {
	let resp = { msg: msg } as any

	if (details) {
		resp.details = details
	}

	return resp
}
