interface JsRunner {
	runjs(code: string): Promise<RunResult>
}

interface RunResult {
	Stdout: Uint8Array
	Stderr: Uint8Array
	ExecTimeMs: number
	ExitCode: number
}

enum lang {
	JS = 'js'
}

class LangNotSupportedError extends Error {
	constructor(requested: string) {
		super(`lang ${requested} not supported`)
		this.name = 'LangNotSupportedError'
	}
}

class Runner {
	js?: JsRunner

	constructor(js?: JsRunner) {
		this.js = js
	}

	public async run(lang: lang, code: string): Promise<RunResult> {
		if (lang == 'js') {
			if (!this.js) {
				throw new LangNotSupportedError(lang)
			}
			return await this.js!.runjs(code)
		}

		throw new LangNotSupportedError(lang)
	}
}

