import { Injectable, NotFoundException } from "@nestjs/common";
import { PrismaService } from "../../prisma/prisma.service";
import { CreateProductDto, UpdateProductDto } from "./products.dto";

@Injectable()
export class ProductsService {
  constructor(private readonly prisma: PrismaService) {}

  async findAll(params: { organizationId: string; search?: string; itemType?: string; skip?: number; take?: number }) {
    const { organizationId, search, itemType, skip = 0, take = 50 } = params;
    return this.prisma.product.findMany({
      where: {
        organizationId,
        ...(itemType ? { itemType: itemType as any } : {}),
        ...(search
          ? {
              OR: [
                { name: { contains: search, mode: "insensitive" } },
                { sku: { contains: search, mode: "insensitive" } },
              ],
            }
          : {}),
      },
      include: { category: true, uom: true },
      skip,
      take,
      orderBy: { name: "asc" },
    });
  }

  async findOne(id: string) {
    const product = await this.prisma.product.findUnique({
      where: { id },
      include: { category: true, uom: true, stockLevels: true },
    });
    if (!product) throw new NotFoundException(`Product ${id} not found`);
    return product;
  }

  async create(organizationId: string, dto: CreateProductDto) {
    return this.prisma.product.create({
      data: { organizationId, ...dto },
    });
  }

  async update(id: string, dto: UpdateProductDto) {
    await this.findOne(id);
    return this.prisma.product.update({ where: { id }, data: dto });
  }

  async remove(id: string) {
    await this.findOne(id);
    // Soft delete preferred over hard delete for audit-trail integrity.
    return this.prisma.product.update({ where: { id }, data: { status: "INACTIVE" } });
  }
}
