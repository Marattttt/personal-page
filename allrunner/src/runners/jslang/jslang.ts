import { spawn } from "child_process"
import { promises } from "fs"
import { join } from "path"
import { JsRunner, RunResult } from "../runner"

export default class Js implements JsRunner {
	private rundir: string

	constructor(dir: string) {
		this.rundir = dir
	}

	async runjs(code: string): Promise<RunResult> {
		const path = join(this.rundir, 'index.js')

		await promises.writeFile(path, code)

		const stdin = `
			cd ${this.rundir};
			npm init -y;
			node index.js;
		`
		return new Promise((resolve, reject) => {
			const cmd = spawn('sh')
			cmd.stdin.write(stdin)
			cmd.stdin.end()

			const res = new RunResult()

			cmd.stdout.on('data', (data) => {
				res.stdout += data.toString()
			})
			cmd.stderr.on('data', (data) => {
				res.stderr += data.toString()
			})

			cmd.on('error', (error) => {
				reject(error)
			})

			cmd.on('close', (code) => {
				if (code === 0) {
					resolve(res)
				} else {
					reject({ error: 'invalid exitcode', exitCode: code })
				}
			})

		})

	}
}
