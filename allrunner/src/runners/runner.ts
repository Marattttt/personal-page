import { Logger } from "pino"
import Js from "./jslang/jslang"

export interface JsRunner {
	runjs(logger: Logger, code: string, timeout: number): Promise<RunResult>
}

export class RunResult {
	stdout: Uint8Array = new Uint8Array()
	stderr: Uint8Array = new Uint8Array()
	execTimeMs: number = -1
	exitCode: number = -1
}

export enum lang {
	JS = 'js'
}

export class LangNotSupportedError extends Error {
	constructor(requested: string) {
		super(`lang ${requested} not supported`)
		this.name = 'LangNotSupportedError'
	}
}

export class Runner {
	private js?: JsRunner

	constructor(js?: JsRunner) {
		this.js = js
	}

	async run(logger: Logger, lang: lang, code: string, timeout: number): Promise<RunResult> {
		if (lang == 'js') {
			if (!this.js) {
				throw new LangNotSupportedError(lang)
			}
			return await this.js!.runjs(logger, code, timeout)
		}

		throw new LangNotSupportedError(lang)
	}
}


export class RunnerBuilder {
	js?: JsRunner

	addJs(dir: string): RunnerBuilder {
		this.js = new Js(dir)
		return this
	}

	bulid(): Runner {
		return new Runner(this.js)
	}
}
