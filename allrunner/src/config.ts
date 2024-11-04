import pino from "pino"
import { HttpLogger, pinoHttp } from "pino-http"

export interface JsConf {
	rundir: string
	timeout: number
}

export interface GoConf {
	rundir: string
	timeout: number
}

export default class AppConfig {
	port: number
	logLevel: pino.Level

	langs: string[]
	jsconf?: JsConf
	goconf?: GoConf

	constructor() {
		this.port = parseInt(process.env.PORT ?? '') || 3000

		this.logLevel = process.env.LOG_LEVEL?.toLowerCase() as pino.Level ?? 'debug'

		const langs = process.env.LANGS;
		if (!langs) {
			throw new Error('LANGS env variable is unset or empty')
		}

		this.langs = langs.split(',').map((l) => l.toUpperCase())
		console.debug({ langs: langs, split: this.langs }, 'processed env var langs')

		if (this.langs.includes('JS')) {
			this.jsconf = {
				rundir: process.env.JS_RUNDIR || './js_rundir',
				timeout:
					parseInt(process.env.JS_TIMEOUT ?? '') ||
					parseInt(process.env.TIMEOUT ?? '') ||
					30_000
			}

			console.debug({ jsconf: this.jsconf }, 'processed js configuration')
		}

		if (this.langs.includes('GO')) {
			this.goconf = {
				rundir: process.env.GO_RUNDIR || './go_rundir',
				timeout:
					parseInt(process.env.GO_TIMEOUT ?? '') ||
					parseInt(process.env.TIMEOUT ?? '') ||
					30_000
			}
		}

		console.info({ conf: this }, 'finished config constructor')
	}

	makeLogger(): HttpLogger {
		return pinoHttp({
			level: this.logLevel,
		})
	}
}
