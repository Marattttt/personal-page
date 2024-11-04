import { Request, Response, Router } from "express";
import { Runner, RunnerBuilder } from "./runners/runner";
import AppConfig, { GoConf, JsConf } from "./config";

export default function createRouter(conf: AppConfig): Router {
	const builder = new RunnerBuilder()
	if (conf.jsconf) {
		builder.addJs(conf.jsconf.rundir)
	}
	if (conf.goconf) {
		builder.addGo(conf.goconf.rundir)
	}

	const runner = builder.bulid()

	const router = Router()

	router.get('/', (_, res) => {
		res.send('<h1> Heeeey </h1>')
	})

	router.post('/js', async (req, res) => {
		conf.jsconf
			? handleRunJs(runner, conf.jsconf, req, res)
			: handleNotSupported('js', req, res)
	})

	router.post('/go', async (req, res) => {
		conf.goconf
			? handleRunGo(runner, conf.goconf, req, res)
			: handleNotSupported('go', req, res)
	})

	return router
}

async function handleNotSupported(lang: string, _: Request, res: Response) {
	res.status(501).json(msgResp(`language ${lang} is not supported`))
}

async function handleRunJs(runner: Runner, conf: JsConf, req: Request, res: Response) {
	let { code } = req.body
	if (!code) {
		res.status(400).json(msgResp('code json body param not provided'))
		return
	}

	code = code!.toString()

	try {
		const runres = await runner.run(
			req.log,
			'js',
			code,
			conf.timeout
		)

		res.status(200).json(runres)
	}
	catch (err) {
		req.log.error({ msg: 'exception during code run', err: err })
		res.status(500).json(msgResp('something went wrong'))
	}
}

async function handleRunGo(runner: Runner, conf: GoConf, req: Request, res: Response) {
	let { code } = req.body
	if (!code) {
		res.status(400).json(msgResp('code json body param not provided'))
		return
	}

	code = code!.toString()

	try {
		const runres = await runner.run(
			req.log,
			'go',
			code,
			conf.timeout
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

