// Script de criação da collection gameservers para LoginServer
// Baseado em: raptors_datapack/sql/gameservers.sql

db = db.getSiblingDB('l2raptors');

// Criar collection gameservers
db.createCollection('gameservers', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['server_id', 'hexid', 'host'],
      properties: {
        server_id: {
          bsonType: 'int',
          description: 'ID unico do GameServer'
        },
        hexid: {
          bsonType: 'string',
          description: 'HexID para autenticacao do GameServer'
        },
        host: {
          bsonType: 'string',
          description: 'Hostname ou IP do GameServer'
        },
        port: {
          bsonType: 'int',
          minimum: 1024,
          maximum: 65535,
          description: 'Porta do GameServer'
        },
        max_players: {
          bsonType: 'int',
          minimum: 0,
          description: 'Numero maximo de jogadores'
        },
        authed: {
          bsonType: 'bool',
          description: 'Se o GameServer esta autenticado'
        },
        pvp: {
          bsonType: 'bool',
          description: 'Se e servidor PvP'
        },
        test_server: {
          bsonType: 'bool',
          description: 'Se e servidor de teste'
        },
        show_clock: {
          bsonType: 'bool',
          description: 'Se mostra relogio'
        },
        show_brackets: {
          bsonType: 'bool',
          description: 'Se mostra colchetes no nome'
        },
        age_limit: {
          bsonType: 'int',
          minimum: 0,
          maximum: 18,
          description: 'Limite de idade'
        },
        server_type: {
          bsonType: 'int',
          description: '0=Normal, 1=Relax, 2=Test, 3=NoLabel, 4=Restricted, 5=Event, 6=Free, 7=World, 8=New, 9=Classic'
        },
        created_at: {
          bsonType: 'date'
        },
        updated_at: {
          bsonType: 'date'
        }
      }
    }
  }
});

// Criar indices
db.gameservers.createIndex({ server_id: 1 }, { unique: true });
db.gameservers.createIndex({ hexid: 1 }, { unique: true });

// Inserir GameServer padrao
db.gameservers.insertOne({
  server_id: 1,
  hexid: '0123456789ABCDEF0123456789ABCDEF',
  host: '127.0.0.1',
  port: 7777,
  max_players: 1000,
  authed: false,
  pvp: true,
  test_server: false,
  show_clock: false,
  show_brackets: true,
  age_limit: 0,
  server_type: 0,
  created_at: new Date(),
  updated_at: new Date()
});

print('Collection gameservers criada com sucesso!');
print('Indices criados: server_id (unique), hexid (unique)');
print('GameServer padrao inserido: ID=1, Port=7777');
