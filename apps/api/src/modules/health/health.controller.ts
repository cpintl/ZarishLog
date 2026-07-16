import { Controller, Get } from "@nestjs/common";
import { ApiTags } from "@nestjs/swagger";

@ApiTags("health")
@Controller("health")
export class HealthController {
  @Get()
  check() {
    return { status: "ok", service: "zarishlog-api", timestamp: new Date().toISOString() };
  }
}
