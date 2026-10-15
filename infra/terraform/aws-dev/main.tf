data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "dev" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_internet_gateway" "dev" {
  vpc_id = aws_vpc.dev.id
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.dev.id
  cidr_block              = "10.42.0.0/24"
  availability_zone       = var.availability_zone
  map_public_ip_on_launch = true
}

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.dev.id
  cidr_block        = "10.42.1.0/24"
  availability_zone = var.availability_zone
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.dev.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.dev.id
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}
