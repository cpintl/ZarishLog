import { Module } from "@nestjs/common";
import { ConfigModule } from "@nestjs/config";
import { PrismaModule } from "./prisma/prisma.module";
import { ProductsModule } from "./modules/products/products.module";
import { HealthController } from "./modules/health/health.controller";

@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true }),
    PrismaModule,
    ProductsModule,
    // Phase 2 will add: WarehousesModule, StockModule, GrnModule, QaModule,
    // AssetsModule, ForecastingModule, ReportingModule, AuthModule.
  ],
  controllers: [HealthController],
})
export class AppModule {}
