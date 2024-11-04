import { exec } from "child_process"
import { promises } from "fs"
import { join } from "path"
import { GoRunner, RunResult, TimeoutMsg as TimeoutMsg } from "../runner"
import AsyncLock from "async-lock"
import { promisify } from "util"
import { Logger } from "pino"

const execPromise = promisify(exec)

const lock = new AsyncLock()
const runGoKey = 'rungo'

export default class Go implements GoRunner {
	private rundir: string

	constructor(dir: string) {
		this.rundir = dir
	}

	/**
	* Runs a main.go file in a directory and returns the output
	* Can throw multiple types of errors
	* Operation is locking across any class instances due to a file-level lock
	*/
	async rungo(logger: Logger, code: string, timeout: number): Promise<RunResult> {
		const res = await lock.acquire(
			runGoKey,
			async () => { return await this.rungoNoLock(logger, code, timeout) }
		)
		return res
	}

	/**
	* The core of the rungo function that is not bound to a lock
	*/
	private async rungoNoLock(logger: Logger, code: string, timeout: number): Promise<RunResult> {
		await prepareDir(this.rundir)

		const path = join(this.rundir, 'main.go')
		await promises.writeFile(path, code)

		logger.info({ content: code, at: path }, 'wrote main.go')

		const res = new RunResult()
		const start = new Date()

		const opts = { cwd: this.rundir, timeout: timeout }

		logger.info({ opts: opts }, 'started execution')

		try {
			const { stdout, stderr } = await execPromise('go run .', opts)

			res.stdout = stdout
			res.stderr = stderr
		}
		catch (error: any) {
			// Node exits with exit code 1 only on an unhandled exception
			// No need to worry about it since this is an error of the submitted code,
			// not the system
			if (error.code > 1) {
				logger.error({
					err: error,
					stdout: error.stdout,
					stderr: error.stderr
				}, 'error running user submitted code')
				throw new Error('Failed go run in runtime dir')
			}

			if (error.killed) {
				error.stderr += TimeoutMsg
				logger.warn('execution timed out')
			}

			res.stdout = error.stdout
			res.stderr = error.stderr

			// exit code might be null
			res.exitCode = error.code ? error.code : 0
		}

		res.execTimeMs = new Date().getTime() - start.getTime()

		logger.info(res, 'finished execution')
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
	go mod init gorunner;
	`
	const proc = exec('sh', (error, stdout, stderr) => {
		if (error) {
			console.error({ msg: 'error creating empty node project', err: error, stdout: stdout, stderr: stderr })
			throw new Error('Could not create empty node project')
		}
	})

	proc.stdin!.write(script)
	proc.stdin!.end()

	await new Promise((resolve) => proc.on('close', resolve))
}
