import { exec } from "child_process"
import { promises } from "fs"
import { join } from "path"
import { JsRunner, RunResult } from "../runner"
import AsyncLock from "async-lock"

const lock = new AsyncLock()
const runjsKey = 'runjs'

export default class Js implements JsRunner {
	private rundir: string

	constructor(dir: string) {
		this.rundir = dir
	}

	/**
	* Runs an index.js file in a directory and returns the output
	* Can throw multiple types of errors
	* Operation is locking across any class instances due to a file-level lock
	*/
	async runjs(code: string, timeout: number): Promise<RunResult> {
		const res = await lock.acquire(
			runjsKey,
			async () => { return await this.runjsNoLock(code, timeout) }
		)
		return res
	}

	/**
	* The core of the runjs function that is not bound to a lock
	*/
	private async runjsNoLock(code: string, timeout: number): Promise<RunResult> {
		await prepareDir(this.rundir)

		const path = join(this.rundir, 'index.js')
		await promises.writeFile(path, code)

		return new Promise((resolve, reject) => {
			const res = new RunResult()
			const start = new Date()

			// Node returns with this exit code on an unhandled exception
			// No need to worry about it since this is an error of the submitted code,
			// not the system
			const exceptionHappened = 1

			const proc = exec('node index.js',
				{
					cwd: this.rundir,
					timeout: timeout
				},
				(error, stdout, stderr) => {
					if (error?.code != exceptionHappened) {
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

/**
 * Creates or clears a directory for executing javascipt 
 * Locks to a file-level lock so concurrent calls are safe
 * Unsafe to use outside of a lock
 */
async function prepareDir(dir: string) {
	// Check if exists and if it does, remove the directory
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
