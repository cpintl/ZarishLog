import { Injectable, OnModuleInit, OnModuleDestroy } from "@nestjs/common";
import { PrismaClient } from "@zarishlog/data-models";

/**
 * Injectable Prisma client. In production, `$use` middleware here also sets
 * the `app.current_org` session variable per-request so Postgres RLS
 * policies (see ARCHITECTURE.md §5) can enforce tenant isolation at the
 * database layer, not just in application code.
 */
@Injectable()
export class PrismaService extends PrismaClient implements OnModuleInit, OnModuleDestroy {
  async onModuleInit() {
    await this.$connect();
  }

  async onModuleDestroy() {
    await this.$disconnect();
  }

  async setTenantContext(organizationId: string) {
    await this.$executeRawUnsafe(`SET app.current_org = '${organizationId}'`);
  }
}
