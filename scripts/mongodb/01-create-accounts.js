// Script de criação da collection accounts para LoginServer
// Baseado em: raptors_datapack/sql/accounts.sql

db = db.getSiblingDB('l2raptors');

// Criar collection accounts
db.createCollection('accounts', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['login', 'password', 'last_active', 'access_level', 'last_server'],
      properties: {
        login: {
          bsonType: 'string',
          description: 'Login do usuario - obrigatorio e unico'
        },
        password: {
          bsonType: 'string',
          description: 'Senha hash BCrypt - obrigatorio'
        },
        last_active: {
          bsonType: 'date',
          description: 'Ultimo acesso - obrigatorio'
        },
        access_level: {
          bsonType: 'int',
          minimum: -100,
          maximum: 100,
          description: 'Nivel de acesso (0=normal, >0=GM, <0=banido)'
        },
        last_server: {
          bsonType: 'int',
          minimum: 1,
          description: 'ID do ultimo servidor usado'
        },
        last_ip: {
          bsonType: 'string',
          description: 'Ultimo IP usado'
        },
        banned_until: {
          bsonType: 'date',
          description: 'Data ate quando esta banido (opcional)'
        },
        created_at: {
          bsonType: 'date',
          description: 'Data de criacao da conta'
        }
      }
    }
  }
});

// Criar indices
db.accounts.createIndex({ login: 1 }, { unique: true });
db.accounts.createIndex({ last_ip: 1 });
db.accounts.createIndex({ access_level: 1 });
db.accounts.createIndex({ banned_until: 1 }, { sparse: true });

print('Collection accounts criada com sucesso!');
print('Indices criados: login (unique), last_ip, access_level, banned_until');
