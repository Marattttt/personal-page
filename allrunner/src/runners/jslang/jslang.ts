import { exec, spawn } from "child_process"
import { promises } from "fs"
import path, { join } from "path"
import { JsRunner, RunResult } from "../runner"

export default class Js implements JsRunner {
	private rundir: string

	constructor(dir: string) {
		this.rundir = dir
	}

	async runjs(code: string): Promise<RunResult> {
		await prepareDir(this.rundir)

		const path = join(this.rundir, 'index.js')
		await promises.writeFile(path, code)

		return new Promise((resolve, reject) => {
			const res = new RunResult()
			const start = new Date()

			const proc = exec('node index.js', { cwd: this.rundir }, (error, stdout, stderr) => {
				if (error) {
					console.error({
						msg: 'error running user submitted code',
						err: error,
						stdout: stdout,
						stderr: stderr
					})
					reject(new Error('Could not run code with nodejs'))
				}

				res.stdout = new TextEncoder().encode(stdout)
				res.stderr = new TextEncoder().encode(stderr)
				res.exitCode = proc.exitCode!
				res.execTimeMs = new Date().getTime() - start.getTime()

				resolve(res)
			})
		})
	}
}

// It is unsafe to execute this function concurrently with the same dir,
// due to possible overrwrites of changes
async function prepareDir(dir: string) {
	// Check if exists and if it does, remove
	try {
		await promises.access(dir)
		await promises.rm(dir, { recursive: true, force: true })
	} catch { }

	await promises.mkdir(dir)

	const script = `
	cd '${dir}';
	npm init -y;
	`

	const proc = exec('sh', (error, stdout, stderr) => {
		if (error) {
			console.error({ msg: 'error creating empty node project', err: error, stdout: stdout, stderr: stderr })
			throw new Error('Could not create empty node project')
		}
	})

	proc.stdin!.write(script)
	proc.stdin!.end()
}
