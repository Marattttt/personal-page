import { exec, ExecException } from "child_process"
import { promises } from "fs"
import { join } from "path"
import { JsRunner, RunResult } from "../runner"
import AsyncLock from "async-lock"
import { promisify } from "util"

const execPromise = promisify(exec)

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

		const res = new RunResult()
		const start = new Date()

		const opts = { cwd: this.rundir, timeout: timeout }

		try {
			const { stdout, stderr } = await execPromise('node index.js', opts)

			res.stdout = new TextEncoder().encode(stdout)
			res.stderr = new TextEncoder().encode(stderr)
		}
		catch (error: any) {
			// Node exits with exit code 1 only on an unhandled exception
			// No need to worry about it since this is an error of the submitted code,
			// not the system
			if (error.code > 1) {
				console.error({
					msg: 'error running user submitted code',
					err: error,
					stdout: error.stdout,
					stderr: error.stderr
				})
				throw new Error('Could not run code with nodejs')
			}

			error.stderr += error.killed ? '\n\nExecution stopped due to timeout' : ''

			res.stdout = new TextEncoder().encode(error.stdout)
			res.stderr = new TextEncoder().encode(error.stderr)
			res.exitCode = error.code ? error.code : 0
		}

		res.execTimeMs = new Date().getTime() - start.getTime()
		return res
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
