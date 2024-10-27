import { Router } from "express";

const router = Router()

router.get('/', (_, res) => {
	res.send('<h1> Heeeey </h1>')
})

export default router
