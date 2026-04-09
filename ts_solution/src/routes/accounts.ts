import { Server } from "@hapi/hapi";
import Joi from "joi";
import { Account } from "../entities/Account";
import AccountService from "../services/account";
import NotFound from "../errors/NotFound";
import InsufficientFunds from "../errors/InsufficientFunds";
import BadRequest from "../errors/BadRequest";
import DuplicateRequest from "../errors/DuplicateRequest";


const validationFailAction = (_request: any, h: any, err: any) => {
  const details = err?.details?.map((d: any) => d.message) ?? [err?.message];
  return h
    .response({ statusCode: 400, error: "Bad Request", message: details })
    .code(400)
    .takeover();
};

const idempotencyKeySchema = Joi.object({
  "idempotency-key": Joi.string().uuid().required(),
}).unknown(true);

export function registerAccountRoutes(server: Server): void {

  server.route({
    method: "GET",
    path: "/accounts/{id}",
    options: {
      validate: {
        params: Joi.object({
          id: Joi.number().integer().required(),
        }),
      },
    },
    handler: async (request, h) => {
      try {
        const account = await AccountService.findById(Number(request.params.id));
        return account;
      } catch (error) {
        if(error instanceof NotFound) {
          return h.response({ error: "Account not found" }).code(404);
        }
        else {
          return h.response({ error }).code(500);
        }
      }
    },
  });

  server.route({
    method: "POST",
    path: "/accounts",
    options: {
      validate: {
        headers: idempotencyKeySchema,
        payload: Joi.object({
          name: Joi.string().min(1).max(255).required(),
          surname: Joi.string().min(1).max(255).required(),
          email: Joi.string().email().max(255).required(),
          phone: Joi.string().max(50).optional().allow(null),
          addressLine1: Joi.string().min(1).max(255).required(),
          addressLine2: Joi.string().max(255).optional().allow(null),
          city: Joi.string().min(1).max(100).required(),
          postcode: Joi.string().min(1).max(20).required(),
          country: Joi.string().min(1).max(100).required(),
          balance: Joi.number().integer().min(0).default(0),
          currency: Joi.string().length(3).uppercase().default("EUR"),
        }),
        failAction: validationFailAction,
      },
    },
    handler: async (request, h) => {
      const idempotencyKey = request.headers["idempotency-key"] as string;

      try {
        const saved = await AccountService.create(idempotencyKey, request.payload as Partial<Account>);
        return h.response(saved).code(201);
      } catch (error) {
        if (error instanceof DuplicateRequest) {
          return h.response(error.cachedResponse).code(error.statusCode);
        }
        return h.response({ error }).code(500);
      }
    },
  });

  server.route({
    method: "POST",
    path: "/accounts/{id}/deposit",
    options: {
      validate: {
        headers: idempotencyKeySchema,
        params: Joi.object({
          id: Joi.number().integer().required(),
        }),
        payload: Joi.object({
          amount: Joi.number().integer().positive().required(),
        }),
        failAction: validationFailAction,
      },
    },
    handler: async (request, h) => {
      const idempotencyKey = request.headers["idempotency-key"] as string;
      const { id } = request.params;
      const { amount } = request.payload as { amount: number };

      try {
        const saved = await AccountService.deposit(idempotencyKey, Number(id), amount);
        return h.response(saved).code(200);
      } catch (error) {
        if (error instanceof DuplicateRequest) {
          return h.response(error.cachedResponse).code(error.statusCode);
        }
        if (error instanceof NotFound) {
          return h.response({ error: "Account not found" }).code(404);
        }
        return h.response({ error }).code(500);
      }
    },
  });

  server.route({
    method: "POST",
    path: "/accounts/{id}/transfer",
    options: {
      validate: {
        headers: idempotencyKeySchema,
        params: Joi.object({
          id: Joi.number().integer().required(),
        }),
        payload: Joi.object({
          toAccountId: Joi.number().integer().required(),
          amount: Joi.number().integer().positive().required(),
        }),
        failAction: validationFailAction,
      },
    },
    handler: async (request, h) => {
      const idempotencyKey = request.headers["idempotency-key"] as string;
      const { id } = request.params;
      const { toAccountId, amount } = request.payload as { toAccountId: number; amount: number };

      try {
        const result = await AccountService.transfer(idempotencyKey, Number(id), toAccountId, amount);
        return h.response(result).code(200);
      } catch (error) {
        if (error instanceof DuplicateRequest) {
          return h.response(error.cachedResponse).code(error.statusCode);
        }
        if (error instanceof BadRequest) {
          return h.response({ statusCode: 400, error: "Bad Request", message: error.message }).code(400);
        }
        if (error instanceof NotFound) {
          return h.response({ error: error.message }).code(404);
        }
        if (error instanceof InsufficientFunds) {
          return h.response({ error: "Insufficient funds" }).code(422);
        }
        return h.response({ error }).code(500);
      }
    },
  });
}
