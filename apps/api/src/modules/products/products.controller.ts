import { Body, Controller, Delete, Get, Headers, Param, Patch, Post, Query } from "@nestjs/common";
import { ApiTags, ApiHeader } from "@nestjs/swagger";
import { ProductsService } from "./products.service";
import { CreateProductDto, UpdateProductDto } from "./products.dto";

/**
 * NOTE: `x-organization-id` header is a placeholder for the tenant context
 * that will come from the authenticated JWT once the auth module (Phase 1,
 * Keycloak/OIDC integration) is wired in. Do not ship this to production
 * without replacing it with a verified claim from the auth guard.
 */
@ApiTags("products")
@ApiHeader({ name: "x-organization-id", required: true })
@Controller("products")
export class ProductsController {
  constructor(private readonly productsService: ProductsService) {}

  @Get()
  findAll(
    @Headers("x-organization-id") organizationId: string,
    @Query("search") search?: string,
    @Query("itemType") itemType?: string,
    @Query("skip") skip?: string,
    @Query("take") take?: string
  ) {
    return this.productsService.findAll({
      organizationId,
      search,
      itemType,
      skip: skip ? Number(skip) : undefined,
      take: take ? Number(take) : undefined,
    });
  }

  @Get(":id")
  findOne(@Param("id") id: string) {
    return this.productsService.findOne(id);
  }

  @Post()
  create(@Headers("x-organization-id") organizationId: string, @Body() dto: CreateProductDto) {
    return this.productsService.create(organizationId, dto);
  }

  @Patch(":id")
  update(@Param("id") id: string, @Body() dto: UpdateProductDto) {
    return this.productsService.update(id, dto);
  }

  @Delete(":id")
  remove(@Param("id") id: string) {
    return this.productsService.remove(id);
  }
}
