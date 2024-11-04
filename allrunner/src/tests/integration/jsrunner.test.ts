import { RunnerBuilder, RunResult, TimeoutMsg } from "../../runners/runner"
import { mkdirSync, rmdirSync } from "fs"
import path from "path"

jest.mock('pino', () => {
	return jest.fn(() => ({
		debug: jest.fn(),
		info: jest.fn(),
		warn: jest.fn(),
		error: jest.fn()
	}))
})
const logger = require('pino')()

const jsdir = path.join(__dirname, 'testdirs', 'jsrunner')

afterAll(() => {
	rmdirSync(jsdir, { recursive: true })
})

test('Output of node is being captured', async () => {
	const builder = new RunnerBuilder()
	builder.addJs(jsdir)
	const runner = builder.bulid()

	const stdout = 'Hello world'
	const stderr = 'some error'

	const code = `
	process.stdout.write('${stdout}');
	process.stderr.write('${stderr}');
	`

	const timeout = 5000

	const result = await runner.run(logger, 'js', code, timeout)

	expect(result).toBeInstanceOf(RunResult)

	expect(result.stdout).toBe(stdout)
	expect(result.stderr).toBe(stderr)
	expect(result.exitCode).toBe(-1)
	expect(result.execTimeMs).toBeLessThanOrEqual(timeout)

})



// test('Produces a timeout error', async () => {
// 	const builder = new RunnerBuilder()
// 	builder.addJs(jsdir)
// 	const runner = builder.bulid()
//
// 	const code = 'await new Promise((resolve) => setTimeout(resolve, 10000))'
//
// 	const result = await runner.run(logger, 'js', code, 1)
//
//
// 	expect(result).toBeInstanceOf(RunResult)
//
// 	expect(result.stderr).toBe(TimeoutMsg)
//
// 	expect(result.stdout).toBe('')
// 	expect(result.exitCode).toBe(0)
// })
