FROM mcr.microsoft.com/dotnet/aspnet:5.0-focal AS base

WORKDIR /app

FROM mcr.microsoft.com/dotnet/sdk:5.0-focal AS build

WORKDIR /src
COPY . .
RUN dotnet restore
RUN dotnet build --no-restore -c Release -o /app

FROM build AS publish
RUN dotnet publish --no-restore -c Release -o /app

FROM base AS final
WORKDIR /app
COPY --from=publish /app .

# Set Hosting Environment to Development
ENV ASPNETCORE_ENVIRONMENT=Development
# Set listening port and wildcare hostname
ENV ASPNETCORE_URLS=http://*:5000

ENTRYPOINT ["dotnet", "dotnet-chain.dll"]