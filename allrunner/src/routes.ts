import { Request, Response, Router } from "express";
import { lang, Runner, RunnerBuilder } from "./runners/runner";

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

	router.post('/js', async (req, res) => {
		handleRunJs(runner, opts, req, res)
	})

	return router
}

async function handleRunJs(runner: Runner, opts: routerOpts, req: Request, res: Response) {
	let { code } = req.body
	if (!code) {
		res.status(400).json(msgResp('code json body param not provided'))
		return
	}

	code = code!.toString()

	try {
		const runres = await runner.run(
			req.log,
			lang.JS,
			code,
			opts.jsTimeout || defJsTimeout
		)

		res.status(200).json(runres)
	}
	catch (err) {
		req.log.error({ msg: 'exception during code run', err: err })
		res.status(500).json(msgResp('something went wrong'))
	}
}

function msgResp(msg: string, details?: object): object {
	const resp = { msg: msg } as any

	if (details) {
		resp.details = details
	}

	return resp
}

