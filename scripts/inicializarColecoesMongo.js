const nomeBanco = db.getName();
const colecoes = [
  'accounts',
  'auctions',
  'augmentations',
  'bbsFavorite',
  'bbsForum',
  'bbsMail',
  'bbsPost',
  'bbsTopic',
  'bookmarks',
  'bufferSchemes',
  'buylists',
  'castle',
  'castleDoorupgrade',
  'castleManorProcure',
  'castleManorProduction',
  'castleTrapupgrade',
  'characterHennas',
  'characterMacroses',
  'characterMemo',
  'characterQuests',
  'characterRaidPoints',
  'characterRecipebook',
  'characterRecommends',
  'characterRelations',
  'characterShortcuts',
  'characterSkills',
  'characterSkillsSave',
  'characterSubclasses',
  'characters',
  'clanData',
  'clanPrivs',
  'clanSkills',
  'clanSubpledges',
  'clanWars',
  'clanhall',
  'clanhallFlagwarAttackers',
  'clanhallFlagwarMembers',
  'clanhallFlagwarOwnerNpcs',
  'clanhallFunctions',
  'clanhallSiegeAttackers',
  'cursedWeapons',
  'fishingChampionship',
  'games',
  'gameservers',
  'grandbossList',
  'heroes',
  'heroesDiary',
  'items',
  'itemsOnGround',
  'mdtBets',
  'mdtHistory',
  'modsWedding',
  'olympiadFights',
  'olympiadNobles',
  'olympiadNoblesEom',
  'petition',
  'petitionMessage',
  'pets',
  'rainbowspringsAttackerList',
  'serverMemo',
  'sevenSigns',
  'sevenSignsFestival',
  'sevenSignsStatus',
  'siegeClans',
  'spawnData'
];

const existentes = new Set(db.getCollectionNames());

for (const nomeColecao of colecoes) {
  if (existentes.has(nomeColecao)) {
    print(`colecao ja existe: ${nomeColecao}`);
    continue;
  }

  db.createCollection(nomeColecao);
  print(`colecao criada: ${nomeColecao}`);
}

print(`inicializacao concluida no banco ${nomeBanco}`);
