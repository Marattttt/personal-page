import { Logger } from "pino"
import Js from "./jslang/jslang"
import Go from "./golang/golang"

export interface GoRunner {
	rungo(loggger: Logger, code: string, timeout: number): Promise<RunResult>
}

export interface JsRunner {
	runjs(logger: Logger, code: string, timeout: number): Promise<RunResult>
}

export class RunResult {
	stdout: string = ''
	stderr: string = ''
	execTimeMs: number = -1
	exitCode: number = -1
}

export type Lang = 'js' | 'go'

export class LangNotSupportedError extends Error {
	constructor(requested: string) {
		super(`lang ${requested} not supported`)
		this.name = 'LangNotSupportedError'
	}
}

export class Runner {
	private js?: JsRunner
	private go?: GoRunner

	constructor(js?: JsRunner, go?: GoRunner) {
		this.js = js
		this.go = go
	}

	async run(logger: Logger, lang: Lang, code: string, timeout: number): Promise<RunResult> {
		if (lang === 'js') {
			if (!this.js) {
				throw new LangNotSupportedError(lang)
			}
			return await this.js.runjs(logger, code, timeout)
		}

		if (lang === 'go') {
			if (!this.go) {
				throw new LangNotSupportedError(lang)
			}
			return await this.go.rungo(logger, code, timeout)
		}

		throw new LangNotSupportedError(lang)
	}
}


export class RunnerBuilder {
	js?: JsRunner
	go?: GoRunner


	addJs(dir: string): RunnerBuilder {
		this.js = new Js(dir)
		return this
	}

	addGo(dir: string): RunnerBuilder {
		this.go = new Go(dir)
		return this
	}

	bulid(): Runner {
		return new Runner(this.js, this.go)
	}
}
