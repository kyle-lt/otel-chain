FROM node:14

# Copy src
WORKDIR /usr/src/app
COPY . .

# Grab libs
RUN npm install

# Rebuild for proper arch
RUN npm rebuild

# Run app
ENTRYPOINT ["node", "-r", "./tracing.js", "app.js"]
