import { IsBoolean, IsEnum, IsOptional, IsString, MinLength } from "class-validator";
import { ApiProperty, PartialType } from "@nestjs/swagger";

export enum ItemTypeDto {
  DRUG = "DRUG",
  MEDICAL_SUPPLY = "MEDICAL_SUPPLY",
  EQUIPMENT = "EQUIPMENT",
  INSTRUMENT = "INSTRUMENT",
  MATERIAL = "MATERIAL",
  VACCINE = "VACCINE",
  NUTRITION = "NUTRITION",
  LAB_REAGENT = "LAB_REAGENT",
  ASSET = "ASSET",
  CONSUMABLE = "CONSUMABLE",
}

export class CreateProductDto {
  @ApiProperty() @IsString() @MinLength(1) sku!: string;
  @ApiProperty() @IsString() @MinLength(1) name!: string;
  @ApiProperty() @IsString() categoryId!: string;
  @ApiProperty() @IsString() uomId!: string;
  @ApiProperty({ enum: ItemTypeDto }) @IsEnum(ItemTypeDto) itemType!: ItemTypeDto;

  @ApiProperty({ required: false }) @IsOptional() @IsString() barcode?: string;
  @ApiProperty({ required: false }) @IsOptional() @IsString() manufacturer?: string;
  @ApiProperty({ required: false, default: false }) @IsOptional() @IsBoolean() batchTracked?: boolean;
  @ApiProperty({ required: false, default: false }) @IsOptional() @IsBoolean() serialTracked?: boolean;
  @ApiProperty({ required: false, default: false }) @IsOptional() @IsBoolean() expiryTracked?: boolean;
  @ApiProperty({ required: false, default: false }) @IsOptional() @IsBoolean() isHazardous?: boolean;
  @ApiProperty({ required: false, default: false }) @IsOptional() @IsBoolean() coldChain?: boolean;
}

export class UpdateProductDto extends PartialType(CreateProductDto) {}
